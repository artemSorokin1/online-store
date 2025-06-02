package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultUserName = "user"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	PassHash  string    `db:"passhash"`
	FullName  string    `db:"fullname"`
	CreatedAt time.Time `db:"created_at"`
	Address   string    `db:"address"`
	City      string    `db:"city"`
	Phone     string    `db:"phone"`
	Role      string    `db:"role"` // "customer" or "seller"
}

type GrpcSeller struct {
	ID        string `json:"id" db:"id"`
	Email     string `json:"email" db:"email"`
	FullName  string `json:"fullname" db:"fullname"`
	CreatedAt string `json:"created_at" db:"created_at"`
	Phone     string `json:"phone" db:"phone"`
	Username  string `json:"username" db:"username"`
}
