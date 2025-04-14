package data

import (
	"book-service/internal/domain"
	"book-service/internal/validator"
	"context"
	"database/sql"
	"errors"
	"time"
)

type FavoriteBookModel struct {
	DB *sql.DB
}

// Insert adds a new favorite book for a user
func (m FavoriteBookModel) Insert(favoriteBook *domain.FavoriteBook) error {
	query := `
		INSERT INTO user_favorite_books (user_id, book_name)
		VALUES ($1, $2)
		RETURNING id, created_at`

	args := []interface{}{favoriteBook.UserID, favoriteBook.BookName}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&favoriteBook.ID, &favoriteBook.CreatedAt)
}

// GetAllForUser retrieves all favorite books for a specific user
func (m FavoriteBookModel) GetAllForUser(userID int64) ([]*domain.FavoriteBook, error) {
	query := `
		SELECT id, user_id, book_name, created_at
		FROM user_favorite_books
		WHERE user_id = $1
		ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favoriteBooks := []*domain.FavoriteBook{}

	for rows.Next() {
		var favoriteBook domain.FavoriteBook
		err := rows.Scan(
			&favoriteBook.ID,
			&favoriteBook.UserID,
			&favoriteBook.BookName,
			&favoriteBook.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		favoriteBooks = append(favoriteBooks, &favoriteBook)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return favoriteBooks, nil
}

// Get retrieves a specific favorite book by its ID
func (m FavoriteBookModel) Get(id int64) (*domain.FavoriteBook, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, user_id, book_name, created_at
		FROM user_favorite_books
		WHERE id = $1`

	var favoriteBook domain.FavoriteBook

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&favoriteBook.ID,
		&favoriteBook.UserID,
		&favoriteBook.BookName,
		&favoriteBook.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &favoriteBook, nil
}

// Delete removes a favorite book by its ID and user ID (for security)
func (m FavoriteBookModel) Delete(id, userID int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM user_favorite_books
		WHERE id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// ValidateFavoriteBook validates the favorite book fields
func ValidateFavoriteBook(v *validator.Validator, favoriteBook *domain.FavoriteBook) {
	v.Check(favoriteBook.UserID != 0, "user_id", "must be provided")
	v.Check(favoriteBook.BookName != "", "book_name", "must be provided")
	v.Check(len(favoriteBook.BookName) <= 255, "book_name", "must not be more than 255 bytes long")
}
