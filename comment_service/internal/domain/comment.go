package domain

import (
	"time"
)

type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProductID uint      `json:"product_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	Rating    int       `json:"rating" gorm:"check:rating >= 1 AND rating <= 5"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 