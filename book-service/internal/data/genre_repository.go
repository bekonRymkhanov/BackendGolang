package data

import (
	"book-service/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type GenreModel struct {
	DB *sql.DB
}

func (e GenreModel) Insert(genre *Genre) error {
	query := `INSERT INTO genres (title, subgenre_count, url)
				VALUES ($1, $2, $3) 
				RETURNING id, version`

	args := []interface{}{genre.Title, genre.SubgenreCount, genre.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return e.DB.QueryRowContext(ctx, query, args...).Scan(&genre.ID, &genre.Version)
}
func (e GenreModel) Get(id int64) (*Genre, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, title, subgenre_count, url, version
				FROM genres
				WHERE id = $1`
	var genre Genre

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&genre.ID,
		&genre.Title,
		&genre.SubgenreCount,
		&genre.URL,
		&genre.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &genre, nil
}

func (e GenreModel) Update(genre *Genre) error {
	query := `UPDATE genres
				SET title = $1, subgenre_count = $2, url = $3, version = version + 1
				WHERE id = $4 AND version = $5
				RETURNING version`

	args := []interface{}{
		genre.Title,
		genre.SubgenreCount,
		genre.URL,
		genre.ID,
		genre.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&genre.Version)
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

func (e GenreModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM genres
				WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := e.DB.ExecContext(ctx, query, id)
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

func (e GenreModel) GetAll(title string, filters Filters) ([]*Genre, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, title, subgenre_count, url, version
		FROM genres
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := e.DB.QueryContext(ctx, query, title, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	genres := []*Genre{}

	for rows.Next() {
		var genre Genre
		err := rows.Scan(
			&totalRecords,
			&genre.ID,
			&genre.Title,
			&genre.SubgenreCount,
			&genre.URL,
			&genre.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		genres = append(genres, &genre)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return genres, metadata, nil
}

func ValidateGenre(v *validator.Validator, genre *Genre) {
	v.Check(genre.Title != "", "title", "must be provided")
	v.Check(genre.SubgenreCount > 0, "sub_genre_count", "must be greater than 0")
	v.Check(len(genre.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(genre.URL != "", "url", "must be provided")

}
