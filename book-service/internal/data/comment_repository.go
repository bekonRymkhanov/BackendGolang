package data

import (
	"book-service/internal/domain"
	"book-service/internal/filters"
	"book-service/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CommentModel struct {
	DB *sql.DB
}

func (m CommentModel) Insert(comment *domain.Comment) error {
	query := `
		INSERT INTO comments (book_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, version
	`

	args := []interface{}{
		comment.BookID,
		comment.UserID,
		comment.Content,
		time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.Version)
}

func (m CommentModel) Get(id int64) (*domain.Comment, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, book_id, user_id, content, created_at, version
		FROM comments
		WHERE id = $1
	`

	var comment domain.Comment

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.BookID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &comment, nil
}

func (m CommentModel) Update(comment *domain.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, version = version + 1
		WHERE id = $2 AND version = $3 AND user_id = $4
		RETURNING version
	`

	args := []interface{}{
		comment.Content,
		comment.ID,
		comment.Version,
		comment.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&comment.Version)
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

func (m CommentModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM comments
		WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
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

func (m CommentModel) GetAllForBook(bookID int64, mfilters filters.Filters) ([]*domain.Comment, filters.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, book_id, user_id, content, created_at, version
		FROM comments
		WHERE book_id = $1
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3
	`, mfilters.SortColumn(), mfilters.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{bookID, mfilters.Limit(), mfilters.Offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	comments := []*domain.Comment{}

	for rows.Next() {
		var comment domain.Comment
		err := rows.Scan(
			&totalRecords,
			&comment.ID,
			&comment.BookID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.Version,
		)
		if err != nil {
			return nil, filters.Metadata{}, err
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, filters.Metadata{}, err
	}

	metadata := filters.CalculateMetadata(totalRecords, mfilters.Page, mfilters.PageSize)

	return comments, metadata, nil
}

func ValidateComment(v *validator.Validator, comment *domain.Comment) {
	v.Check(comment.BookID > 0, "book_id", "must be provided")
	v.Check(comment.UserID > 0, "user_id", "must be provided")
	v.Check(comment.Content != "", "content", "must be provided")
	v.Check(len(comment.Content) <= 1000, "content", "must not be more than 1000 bytes long")
}
