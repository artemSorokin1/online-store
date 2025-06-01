package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Product struct {
	ID          int64           `db:"id"`
	Name        string          `db:"name"`
	Description sql.NullString  `db:"description"`
	Price       int64           `db:"price"`
	ImageURL    sql.NullString  `db:"image_url"`
	Info        json.RawMessage `db:"info"`
	SellerID    uuid.UUID       `db:"seller_id"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
	Rating      float64         `db:"rating"`
	Comments    []string        `db:"comments"`
}

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetProduct(ctx context.Context, id int64) (*Product, error) {
	query := `
		WITH product_ratings AS (
			SELECT 
				product_id,
				AVG(value) as avg_rating
			FROM product_ratings
			GROUP BY product_id
		)
		SELECT 
			p.id,
			p.name,
			p.description,
			p.price,
			p.image_url,
			p.info,
			p.seller_id,
			p.created_at,
			p.updated_at,
			COALESCE(pr.avg_rating, 0) as rating,
			ARRAY_AGG(DISTINCT c.content) as comments
		FROM products p
		LEFT JOIN product_ratings pr ON p.id = pr.product_id
		LEFT JOIN comments c ON p.id = c.product_id
		WHERE p.id = $1
		GROUP BY p.id, pr.avg_rating`

	var product Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.ImageURL,
		&product.Info,
		&product.SellerID,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Rating,
		pq.Array(&product.Comments),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (
			name, description, price, image_url, info, seller_id
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.Info,
		product.SellerID,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *Product) error {
	query := `
		UPDATE products
		SET 
			name = $1,
			description = $2,
			price = $3,
			image_url = $4,
			info = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.Info,
		product.ID,
	).Scan(&product.UpdatedAt)
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
} 