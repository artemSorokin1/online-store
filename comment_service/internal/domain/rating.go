package domain

import (
	"time"
)

type Rating struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProductID uint      `json:"product_id"`
	UserID    uint      `json:"user_id"`
	Value     int       `json:"value" gorm:"check:value >= 1 AND value <= 5"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 