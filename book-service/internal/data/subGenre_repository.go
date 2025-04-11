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

type SubGenreModel struct {
	DB *sql.DB
}

func (e SubGenreModel) Insert(sub_genre *domain.SubGenre) error {
	query := `INSERT INTO subgenres ( title, main_genre, book_count, url)
				VALUES ($1, $2, $3, $4)
				RETURNING id, version`

	args := []interface{}{sub_genre.Title, sub_genre.MainGenre, sub_genre.BookCount, sub_genre.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return e.DB.QueryRowContext(ctx, query, args...).Scan(&sub_genre.ID, &sub_genre.Version)
}

func (m SubGenreModel) Get(id int64) (*domain.SubGenre, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, title, main_genre, book_count, url, 1 FROM subgenres WHERE id = $1`

	var subGenre domain.SubGenre

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&subGenre.ID,
		&subGenre.Title,
		&subGenre.MainGenre,
		&subGenre.BookCount,
		&subGenre.URL,
		&subGenre.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &subGenre, nil
}

func (m SubGenreModel) Update(subGenre *domain.SubGenre) error {
	query := `UPDATE subgenres
				SET title = $1, main_genre = $2, book_count = $3, url = $4
				WHERE id = $5 AND version = $6
				RETURNING 1`

	args := []interface{}{
		subGenre.Title,
		subGenre.MainGenre,
		subGenre.BookCount,
		subGenre.URL,
		subGenre.ID,
		subGenre.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&subGenre.Version)
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

func (m SubGenreModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM subgenres WHERE id = $1`

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

func (m SubGenreModel) GetAll(title string, mfilters filters.Filters) ([]*domain.SubGenre, filters.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, title, main_genre, book_count, url, 1
		FROM subgenres
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, mfilters.SortColumn(), mfilters.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, mfilters.Limit(), mfilters.Offset())
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	subGenres := []*domain.SubGenre{}

	for rows.Next() {
		var sg domain.SubGenre
		err := rows.Scan(
			&totalRecords,
			&sg.ID,
			&sg.Title,
			&sg.MainGenre,
			&sg.BookCount,
			&sg.URL,
			&sg.Version,
		)

		if err != nil {
			return nil, filters.Metadata{}, err
		}
		subGenres = append(subGenres, &sg)
	}

	if err = rows.Err(); err != nil {
		return nil, filters.Metadata{}, err
	}

	metadata := filters.CalculateMetadata(totalRecords, mfilters.Page, mfilters.PageSize)
	return subGenres, metadata, nil
}

func (m SubGenreModel) GetByGenre(genre string) ([]*domain.SubGenre, error) {
	query := `SELECT id, title, main_genre, book_count, url, 1 FROM subgenres WHERE main_genre = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, genre)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subGenres := []*domain.SubGenre{}

	for rows.Next() {
		var sg domain.SubGenre
		err := rows.Scan(
			&sg.ID,
			&sg.Title,
			&sg.MainGenre,
			&sg.BookCount,
			&sg.URL,
			&sg.Version,
		)
		if err != nil {
			return nil, err
		}
		subGenres = append(subGenres, &sg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subGenres, nil
}

func ValidateSubGenre(v *validator.Validator, subgenre *domain.SubGenre) {
	v.Check(subgenre.Title != "", "title", "must be provided")
	v.Check(subgenre.MainGenre != "", "main_genre", "must be provided")
	v.Check(len(subgenre.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(subgenre.BookCount > 0, "book_count", "must be greater than 0")
	v.Check(subgenre.URL != "", "url", "must be provided")
}
