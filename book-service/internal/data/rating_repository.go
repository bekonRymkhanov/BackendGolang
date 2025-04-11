package data

import (
	"book-service/internal/validator"
	"context"
	"database/sql"
	"errors"
	"time"
)

type RatingModel struct {
	DB *sql.DB
}

func (m RatingModel) Insert(rating *Rating) error {
	// First check if the user has already rated this book
	query := `
		SELECT id, version FROM ratings
		WHERE book_id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var existingID int64
	var existingVersion int32

	err := m.DB.QueryRowContext(ctx, query, rating.BookID, rating.UserID).Scan(&existingID, &existingVersion)

	if err == nil {
		rating.ID = existingID
		rating.Version = existingVersion
		return m.Update(rating)
	}

	if errors.Is(err, sql.ErrNoRows) {
		insertQuery := `
			INSERT INTO ratings (book_id, user_id, score, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, version, created_at
		`

		args := []interface{}{
			rating.BookID,
			rating.UserID,
			rating.Score,
			time.Now(),
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel2()

		return m.DB.QueryRowContext(ctx2, insertQuery, args...).Scan(
			&rating.ID,
			&rating.Version,
			&rating.CreatedAt,
		)
	}

	return err
}

func (m RatingModel) Get(id int64) (*Rating, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, book_id, user_id, score, created_at, version
		FROM ratings
		WHERE id = $1
	`

	var rating Rating

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&rating.ID,
		&rating.BookID,
		&rating.UserID,
		&rating.Score,
		&rating.CreatedAt,
		&rating.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &rating, nil
}

func (m RatingModel) GetUserRatingForBook(userID, bookID int64) (*Rating, error) {
	query := `
		SELECT id, book_id, user_id, score, created_at, version
		FROM ratings
		WHERE user_id = $1 AND book_id = $2
	`

	var rating Rating

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID, bookID).Scan(
		&rating.ID,
		&rating.BookID,
		&rating.UserID,
		&rating.Score,
		&rating.CreatedAt,
		&rating.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &rating, nil
}

func (m RatingModel) Update(rating *Rating) error {
	query := `
		UPDATE ratings
		SET score = $1, version = version + 1
		WHERE id = $2 AND version = $3 AND user_id = $4
		RETURNING version
	`

	args := []interface{}{
		rating.Score,
		rating.ID,
		rating.Version,
		rating.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&rating.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m RatingModel) Delete(id, userID int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM ratings
		WHERE id = $1 AND user_id = $2
	`

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

func (m RatingModel) GetAllForBook(bookID int64) ([]*Rating, error) {
	query := `
		SELECT id, book_id, user_id, score, created_at, version
		FROM ratings
		WHERE book_id = $1
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := []*Rating{}

	for rows.Next() {
		var rating Rating
		err := rows.Scan(
			&rating.ID,
			&rating.BookID,
			&rating.UserID,
			&rating.Score,
			&rating.CreatedAt,
			&rating.Version,
		)
		if err != nil {
			return nil, err
		}

		ratings = append(ratings, &rating)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ratings, nil
}

func (m RatingModel) GetAverageRating(bookID int64) (float64, int, error) {
	query := `
		SELECT COALESCE(AVG(score), 0) as average_score, COUNT(*) as rating_count
		FROM ratings
		WHERE book_id = $1
	`

	var averageScore float64
	var count int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, bookID).Scan(&averageScore, &count)
	if err != nil {
		return 0, 0, err
	}

	return averageScore, count, nil
}

func ValidateRating(v *validator.Validator, rating *Rating) {
	v.Check(rating.BookID > 0, "book_id", "must be provided")
	v.Check(rating.UserID > 0, "user_id", "must be provided")
	v.Check(rating.Score >= 1 && rating.Score <= 5, "score", "must be between 1 and 5")
}
