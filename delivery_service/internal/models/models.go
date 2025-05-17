package models

import (
	"github.com/lib/pq"
	"time"
)

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

type Product struct {
	ID          int           `db:"id"`
	Name        string        `db:"name"`
	Price       int           `db:"price"`
	Sizes       pq.Int64Array `db:"sizes"`
	ImageURL    string        `db:"imageurl"`
	Description string        `db:"description"`
}

type Order struct {
	ProductIds   []int
	OrderId      int
	CreationTime time.Duration
}

type Courier struct {
	CourierID int
	Sex       string
}

type CartItem struct {
	ProductId int    `db:"product_id" json:"productId"`
	Name      string `db:"name" json:"name"`
	Price     int    `db:"price" json:"price"`
	Size      int    `db:"size" json:"size"`
	ImageURL  string `db:"imageurl" json:"imageURL"`
}

type CartInfo struct {
	ProductId int64 `db:"product_id"`
	Size      int64 `db:"size"`
}
