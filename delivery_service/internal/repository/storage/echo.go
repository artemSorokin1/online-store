package storage

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/models"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// так надо
//type Storage interface {
//
//}

type DB struct {
	Db *sqlx.DB
}

func New(config *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", config.DBCfg.Host, config.DBCfg.Username, config.DBCfg.Password, config.DBCfg.DBName, config.DBCfg.Port)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := db.Conn(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	return &DB{Db: db}, nil
}

func (d *DB) GetUserById(userId int64) (models.User, error) {
	var user models.User
	err := d.Db.Get(&user, "SELECT * FROM users WHERE id=$1", userId)
	if err != nil {
		return models.User{}, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

func (d *DB) GetProducts() ([]models.Product, error) {
	var products []models.Product
	err := d.Db.Select(&products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("error getting products: %w", err)
	}

	return products, nil
}

func (d *DB) GetProductById(id int64) (models.Product, error) {
	var product models.Product
	err := d.Db.Get(&product, "SELECT * FROM products WHERE id=$1", id)
	if err != nil {
		return models.Product{}, fmt.Errorf("error getting product: %w", err)
	}

	return product, nil
}

func (d *DB) CreateCart(userId int64) error {
	_, err := d.Db.Exec("INSERT INTO cart (user_id, total) VALUES ($1, $2)", userId, 0)
	if err != nil {
		return fmt.Errorf("error creating cart: %w", err)
	}

	return nil
}

func (d *DB) GetCart(userId int64) ([]models.CartItem, error) {
	query := `SELECT product_id, name, price, size, products.imageurl
				FROM
					cart_items
				JOIN
						cart
				ON cart_items.cart_id = cart.id
				JOIN
						products
				ON cart_items.product_id = products.id
				WHERE cart.user_id = $1;`

	var cart []models.CartItem

	err := d.Db.Select(&cart, query, userId)
	if err != nil {
		slog.Warn("error getting cart: %w", err)
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	return cart, nil
}

func (d *DB) AddProductInCart(userId int64, productId int64, size int64, price int64) error {
	var cartId int64

	err := d.Db.Get(&cartId, "SELECT id FROM cart WHERE user_id=$1", userId)
	if err != nil {
		return fmt.Errorf("error getting cart: %w", err)
	}

	_, err = d.Db.Exec("INSERT INTO cart_items (cart_id, product_id, size) VALUES ($1, $2, $3)", cartId, productId, size)
	if err != nil {
		return fmt.Errorf("error adding product in cart: %w", err)
	}

	_, err = d.Db.Exec("UPDATE cart SET total = total + $1 WHERE id = $2", price, cartId)
	if err != nil {
		return fmt.Errorf("error updating cart: %w", err)
	}

	return nil
}

func (d *DB) DeleteCart(userId int64) error {
	query := `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM cart WHERE user_id = $1)`
	_, err := d.Db.Exec(query, userId)
	if err != nil {
		return fmt.Errorf("error deleting cart items: %w", err)
	}

	_, err = d.Db.Exec("UPDATE cart SET total = 0 WHERE user_id = $1", userId)
	if err != nil {
		return fmt.Errorf("error updating total: %w", err)
	}

	return nil
}

func (d *DB) DeleteCartItem(userId int64, productId int64) error {
	query := `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM cart WHERE user_id = $1) AND product_id = $2`
	_, err := d.Db.Exec(query, userId, productId)
	if err != nil {
		return fmt.Errorf("error deleting cart item: %w", err)
	}

	query = `UPDATE cart SET total = total - (SELECT price FROM products WHERE id = $2) WHERE user_id = $1`
	_, err = d.Db.Exec(query, userId, productId)
	if err != nil {
		return fmt.Errorf("error updating total: %w", err)
	}

	return nil
}

func (d *DB) GetCartInfo(userId int64) ([]models.CartInfo, error) {
	var cartInfo []models.CartInfo

	query := `SELECT product_id, size 
				FROM
					cart_items
				JOIN
					cart
				ON cart_items.cart_id = cart.id
				WHERE cart.user_id = $1
				    `

	err := d.Db.Select(&cartInfo, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting product ids: %w", err)
	}

	return cartInfo, nil
}

func (d *DB) GetTotalPrice(userId int64) (int, error) {
	var totalPrice int

	query := `SELECT total FROM cart WHERE user_id = $1`

	err := d.Db.Get(&totalPrice, query, userId)
	if err != nil {
		return 0, fmt.Errorf("error getting total price: %w", err)
	}

	return totalPrice, nil
}

func (d *DB) CreateOrder(userId int64, name string, address string, phone string, productIds pq.Int64Array, sizes pq.Int64Array, total int) error {
	query := `
		INSERT INTO orders (user_id, productids, sizes, total, name, address, userphone, orderdate) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`

	_, err := d.Db.Exec(query, userId, productIds, sizes, total, name, address, phone)
	if err != nil {
		fmt.Println("error creating order: ", err)
		return fmt.Errorf("error creating order: %w", err)
	}

	return nil
}

func (d *DB) GetOrderId(userId int64) (int, error) {
	q := `SELECT id FROM orders WHERE user_id = $1 ORDER BY orderdate DESC LIMIT 1`

	var orderId int
	err := d.Db.Get(&orderId, q, userId)
	if err != nil {
		return 0, fmt.Errorf("error getting order id: %w", err)
	}

	return orderId, nil
}

func (d *DB) AddOrderInUserActions(userId int64, orderId int) error {
	query := `INSERT INTO user_actions (user_id, order_id, action, created_at) VALUES ($1, $2, 'created_order', NOW())`

	_, err := d.Db.Exec(query, userId, orderId)
	if err != nil {
		return fmt.Errorf("error adding order in user actions: %w", err)
	}

	return nil
}

func (d *DB) AdminAddProduct(product models.Product) error {
	query := `
		INSERT INTO products (name, price, sizes, imageurl, description)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := d.Db.Exec(query, product.Name, product.Price, product.Sizes, product.ImageURL, product.Description)
	if err != nil {
		fmt.Println("error adding product: ", err)
		return fmt.Errorf("error adding product: %w", err)
	}

	return nil
}

func (d *DB) AdminDeleteProduct(productId int64) error {
	_, err := d.Db.Exec("DELETE FROM products WHERE id = $1", productId)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

func (d *DB) GetLastInsertedProduct() (models.Product, error) {
	var product models.Product

	err := d.Db.Get(&product, "SELECT * FROM products ORDER BY id DESC LIMIT 1")
	if err != nil {
		return models.Product{}, fmt.Errorf("error getting last inserted product id: %w", err)
	}

	return product, nil
}
