package storage

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/models"
	"fmt"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	Stage = "repository"
)

// так надо
//type Storage interface {
//
//}

type DB struct {
	Db     *sqlx.DB
	logger *zap.Logger
}

func New(config *config.Config, logger *zap.Logger) (*DB, error) {
	logger.Info("connecting to db")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.DBCfg.Host,
		config.DBCfg.Username,
		config.DBCfg.Password,
		config.DBCfg.DBName,
		config.DBCfg.Port,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.Fatal("unable to connect to db", zap.Error(err))
	}
	if _, err := db.Conn(context.Background()); err != nil {
		logger.Fatal("unable to connect to db", zap.Error(err))
	}

	logger.Info("storage run")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		logger.Fatal("error creating postgres driver", zap.Error(err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		logger.Fatal("error creating new migration instance", zap.Error(err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal("error running migrations", zap.Error(err))
	}

	logger.Info("migrations successfully applied")

	return &DB{
		Db:     db,
		logger: logger,
	}, nil
}

func (d *DB) GetUserById(userId int64) (models.User, error) {
	d.logger.Debug("getting user by id", zap.Int64("userId", userId))

	var user models.User
	err := d.Db.Get(&user, "SELECT * FROM users WHERE id=$1", userId)
	if err != nil {
		d.logger.Error("error getting user by id")
		return models.User{}, fmt.Errorf("error getting user: %w", err)
	}

	d.logger.Debug("successfully got user by id", zap.Int64("userId", userId))

	return user, nil
}

func (d *DB) GetProducts() ([]models.Product, error) {
	d.logger.Debug("getting products")
	var products []models.Product
	err := d.Db.Select(&products, "SELECT * FROM products")
	if err != nil {
		d.logger.Error("error getting products")
		return nil, fmt.Errorf("error getting products: %w", err)
	}

	d.logger.Debug("successfully got products")

	return products, nil
}

func (d *DB) GetProductById(id int64) (models.Product, error) {
	d.logger.Debug("getting product by id", zap.Int64("productId", id))
	var product models.Product
	err := d.Db.Get(&product, "SELECT * FROM products WHERE id=$1", id)
	if err != nil {
		d.logger.Error("error getting product by id")
		return models.Product{}, fmt.Errorf("error getting product: %w", err)
	}

	d.logger.Debug("successfully got product by id", zap.Int64("productId", id))

	return product, nil
}

func (d *DB) CreateCart(userId int64) error {
	d.logger.Debug("creating cart for user", zap.Int64("userId", userId))

	_, err := d.Db.Exec("INSERT INTO cart (user_id, total) VALUES ($1, $2)", userId, 0)
	if err != nil {
		d.logger.Error("error creating cart")
		return fmt.Errorf("error creating cart: %w", err)
	}

	d.logger.Debug("successfully created cart for user", zap.Int64("userId", userId))

	return nil
}

func (d *DB) GetCart(userId int64) ([]models.CartItem, error) {
	d.logger.Debug("getting cart for user", zap.Int64("userId", userId))

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
		d.logger.Error("error getting cart")
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	d.logger.Debug("successfully got cart for user", zap.Int64("userId", userId))

	return cart, nil
}

// TODO: переделать
func (d *DB) AddProductInCart(userId int64, productId int64, size int64, price int64) error {
	d.logger.Debug("adding product in cart", zap.Int64("userId", userId), zap.Int64("productId", productId))
	var cartId int64

	err := d.Db.Get(&cartId, "SELECT id FROM cart WHERE user_id=$1", userId)
	if err != nil {
		d.logger.Error("error getting cart")
		return fmt.Errorf("error getting cart: %w", err)
	}

	_, err = d.Db.Exec("INSERT INTO cart_items (cart_id, product_id, size) VALUES ($1, $2, $3)", cartId, productId, size)
	if err != nil {
		d.logger.Error("error adding product in cart")
		return fmt.Errorf("error adding product in cart: %w", err)
	}

	_, err = d.Db.Exec("UPDATE cart SET total = total + $1 WHERE id = $2", price, cartId)
	if err != nil {
		d.logger.Error("error updating cart")
		return fmt.Errorf("error updating cart: %w", err)
	}

	d.logger.Debug("successfully added product in cart", zap.Int64("userId", userId), zap.Int64("productId", productId))

	return nil
}

func (d *DB) DeleteCart(userId int64) error {
	d.logger.Debug("deleting cart for user", zap.Int64("userId", userId))

	query := `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM cart WHERE user_id = $1)`
	_, err := d.Db.Exec(query, userId)
	if err != nil {
		d.logger.Error("error deleting cart items")
		return fmt.Errorf("error deleting cart items: %w", err)
	}

	_, err = d.Db.Exec("UPDATE cart SET total = 0 WHERE user_id = $1", userId)
	if err != nil {
		d.logger.Error("error updating cart total")
		return fmt.Errorf("error updating total: %w", err)
	}

	d.logger.Debug("successfully deleted cart for user", zap.Int64("userId", userId))

	return nil
}

func (d *DB) DeleteCartItem(userId int64, productId int64) error {
	d.logger.Debug("deleting cart item", zap.Int64("userId", userId), zap.Int64("productId", productId))

	query := `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM cart WHERE user_id = $1) AND product_id = $2`
	_, err := d.Db.Exec(query, userId, productId)
	if err != nil {
		d.logger.Error("error deleting cart item")
		return fmt.Errorf("error deleting cart item: %w", err)
	}

	query = `UPDATE cart SET total = total - (SELECT price FROM products WHERE id = $2) WHERE user_id = $1`
	_, err = d.Db.Exec(query, userId, productId)
	if err != nil {
		d.logger.Error("error updating cart total")
		return fmt.Errorf("error updating total: %w", err)
	}

	d.logger.Debug("successfully deleted cart item", zap.Int64("userId", userId), zap.Int64("productId", productId))

	return nil
}

func (d *DB) GetCartInfo(userId int64) ([]models.CartInfo, error) {
	d.logger.Debug("getting cart info", zap.Int64("userId", userId))

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
		d.logger.Error("error getting product ids")
		return nil, fmt.Errorf("error getting product ids: %w", err)
	}

	d.logger.Debug("successfully got cart info", zap.Int64("userId", userId))

	return cartInfo, nil
}

func (d *DB) GetTotalPrice(userId int64) (int, error) {
	d.logger.Debug("getting total price", zap.Int64("userId", userId))
	var totalPrice int

	query := `SELECT total FROM cart WHERE user_id = $1`

	err := d.Db.Get(&totalPrice, query, userId)
	if err != nil {
		d.logger.Error("error getting total price")
		return 0, fmt.Errorf("error getting total price: %w", err)
	}

	d.logger.Debug("successfully got total price", zap.Int64("userId", userId), zap.Int("totalPrice", totalPrice))

	return totalPrice, nil
}

// TODO: наговнокодил - переделать
func (d *DB) CreateOrder(userId int64, name string, address string, phone string, productIds pq.Int64Array, sizes pq.Int64Array, total int) error {
	d.logger.Debug("creating order", zap.Int64("userId", userId), zap.String("name", name), zap.String("address", address), zap.String("phone", phone))

	query := `
		INSERT INTO orders (user_id, productids, sizes, total, name, address, userphone, orderdate) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`

	_, err := d.Db.Exec(query, userId, productIds, sizes, total, name, address, phone)
	if err != nil {
		d.logger.Error("error creating order")
		return fmt.Errorf("error creating order: %w", err)
	}

	d.logger.Debug("successfully created order", zap.Int64("userId", userId), zap.String("name", name), zap.String("address", address), zap.String("phone", phone))

	return nil
}

func (d *DB) GetOrderId(userId int64) (int, error) {
	d.logger.Debug("getting order id", zap.Int64("userId", userId))

	q := `SELECT id FROM orders WHERE user_id = $1 ORDER BY orderdate DESC LIMIT 1`

	var orderId int
	err := d.Db.Get(&orderId, q, userId)
	if err != nil {
		d.logger.Error("error getting order id")
		return 0, fmt.Errorf("error getting order id: %w", err)
	}

	d.logger.Debug("successfully got order id", zap.Int64("userId", userId), zap.Int("orderId", orderId))

	return orderId, nil
}

func (d *DB) AddOrderInUserActions(userId int64, orderId int) error {
	d.logger.Debug("adding order in user actions", zap.Int64("userId", userId), zap.Int("orderId", orderId))

	query := `INSERT INTO user_actions (user_id, order_id, action, created_at) VALUES ($1, $2, 'created_order', NOW())`

	_, err := d.Db.Exec(query, userId, orderId)
	if err != nil {
		d.logger.Error("error adding order in user actions")
		return fmt.Errorf("error adding order in user actions: %w", err)
	}

	d.logger.Debug("successfully added order in user actions", zap.Int64("userId", userId), zap.Int("orderId", orderId))

	return nil
}

func (d *DB) AdminAddProduct(product models.Product) error {
	d.logger.Debug("adding product", zap.String("productName", product.Name))

	query := `
		INSERT INTO products (name, price, sizes, imageurl, description)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := d.Db.Exec(query, product.Name, product.Price, product.Sizes, product.ImageURL, product.Description)
	if err != nil {
		d.logger.Error("error adding product")
		fmt.Println("error adding product: ", err)
		return fmt.Errorf("error adding product: %w", err)
	}

	d.logger.Debug("successfully added product", zap.String("productName", product.Name))

	return nil
}

func (d *DB) AdminDeleteProduct(productId int64) error {
	d.logger.Debug("deleting product", zap.Int64("productId", productId))

	_, err := d.Db.Exec("DELETE FROM products WHERE id = $1", productId)
	if err != nil {
		d.logger.Error("error deleting product")
		return fmt.Errorf("error deleting product: %w", err)
	}

	d.logger.Debug("successfully deleted product", zap.Int64("productId", productId))

	return nil
}

func (d *DB) GetLastInsertedProduct() (models.Product, error) {
	d.logger.Debug("getting last inserted product")

	var product models.Product

	err := d.Db.Get(&product, "SELECT * FROM products ORDER BY id DESC LIMIT 1")
	if err != nil {
		d.logger.Error("error getting last inserted product")
		return models.Product{}, fmt.Errorf("error getting last inserted product id: %w", err)
	}

	d.logger.Debug("successfully got last inserted product", zap.Int("productId", product.ID))

	return product, nil
}
