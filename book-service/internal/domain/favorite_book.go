package domain

import (
	"time"
)

// FavoriteBook represents a user's favorite book
type FavoriteBook struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	BookName  string    `json:"book_name"`
	CreatedAt time.Time `json:"created_at"`
}
