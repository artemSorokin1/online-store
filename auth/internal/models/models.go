package models

import "time"

const (
	DefaultUserName = "user"
)

type User struct {
	ID        int       `db:"id"`
	Email     string    `db:"email"`
	PassHash  string    `db:"passhash"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"created_at"`
}

type RefreshToken struct {
	ID        int       `db:"id"`
	TokenHash string    `db:"token_hash"`
	userId    int       `db:"user_id"`
	expiresAt time.Time `db:"expires_at"`
}
