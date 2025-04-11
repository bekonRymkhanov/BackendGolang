package data

import "time"

type Rating struct {
	ID        int64     `json:"id"`
	BookID    int64     `json:"book_id"`
	UserID    int64     `json:"user_id"`
	Score     int       `json:"score"` // Rating score, e.g., 1-5
	CreatedAt time.Time `json:"created_at"`
	Version   int32     `json:"version"`
}
