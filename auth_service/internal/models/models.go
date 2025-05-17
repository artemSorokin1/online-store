package models

import "time"

type User struct {
	ID             int64     `db:"id"`
	Email          string    `db:"email"`
	Username       string    `db:"username"`
	PassHash       string    `db:"passhash"`
	TimeCreatedAcc time.Time `db:"created_acc"`
	Role           string    `db:"role"`
}
