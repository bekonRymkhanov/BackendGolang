package data

import (
	"book-service/internal/domain"
	"book-service/internal/filters"
	"book-service/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type BookModel struct {
	DB *sql.DB
}

func (e BookModel) Insert(book *domain.Book) error {
	query := `INSERT INTO books (title, author, main_genre, sub_genre, type, price, rating, people_rated, url) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id, version`

	args := []interface{}{book.Author, book.Title, book.MainGenre, book.SubGenre, book.Type, book.Price, book.Rating, book.PeopleRated, book.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return e.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID, &book.Version)
}
func (e BookModel) Get(id int64) (*domain.Book, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id,author, title, main_genre, sub_genre, type, price, rating, people_rated, url, version
				FROM books
				WHERE id = $1`
	var book domain.Book

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Author,
		&book.Title,
		&book.MainGenre,
		&book.SubGenre,
		&book.Type,
		&book.Price,
		&book.Rating,
		&book.PeopleRated,
		&book.URL,
		&book.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &book, nil
}

func (e BookModel) Update(book *domain.Book) error {
	query := `UPDATE books
				SET title = $1, author = $2, main_genre = $3, sub_genre = $4, type = $5, price = $6, rating = $7, people_rated = $8, url = $9, version = version + 1
				WHERE id = $10 AND version = $11
				RETURNING version`

	args := []interface{}{
		book.Title,
		book.Author,
		book.MainGenre,
		book.SubGenre,
		book.Type,
		book.Price,
		book.Rating,
		book.PeopleRated,
		book.URL,
		book.ID,
		book.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&book.Version)
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

func (e BookModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM books
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

func (e BookModel) GetAll(mfilters filters.Filters, msearchOptions filters.BookSearch) ([]*domain.Book, filters.Metadata, error) {
	where := []string{}
	args := []interface{}{}
	argPosition := 1

	if msearchOptions.Title != "" {
		where = append(where, fmt.Sprintf(
			"(to_tsvector('simple', title) @@ plainto_tsquery('simple', $%d) OR title ILIKE $%d)", argPosition, argPosition+1))
		args = append(args, msearchOptions.Title, "%"+msearchOptions.Title+"%")
		argPosition += 2
	}

	if msearchOptions.Author != "" {
		where = append(where, fmt.Sprintf(
			"(to_tsvector('simple', author) @@ plainto_tsquery('simple', $%d) OR author ILIKE $%d)", argPosition, argPosition+1))
		args = append(args, msearchOptions.Author, "%"+msearchOptions.Author+"%")
		argPosition += 2
	}
	if msearchOptions.Main_genre != "" {
		where = append(where, fmt.Sprintf(
			"(to_tsvector('simple', main_genre) @@ plainto_tsquery('simple', $%d) OR main_genre ILIKE $%d)", argPosition, argPosition+1))
		args = append(args, msearchOptions.Main_genre, "%"+msearchOptions.Main_genre+"%")
		argPosition += 2
	}
	if msearchOptions.Sub_genre != "" {
		where = append(where, fmt.Sprintf(
			"(to_tsvector('simple', sub_genre) @@ plainto_tsquery('simple', $%d) OR sub_genre ILIKE $%d)", argPosition, argPosition+1))
		args = append(args, msearchOptions.Sub_genre, "%"+msearchOptions.Sub_genre+"%")
		argPosition += 2
	}
	if msearchOptions.Type != "" {
		where = append(where, fmt.Sprintf(
			"(to_tsvector('simple', type) @@ plainto_tsquery('simple', $%d) OR type ILIKE $%d)", argPosition, argPosition+1))
		args = append(args, msearchOptions.Type, "%"+msearchOptions.Type+"%")
		argPosition += 2
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " OR ")
	}

	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, author, title, main_genre, sub_genre, type, price, rating, people_rated, url, version
	FROM books
	%s
	ORDER BY %s %s, id ASC 
	LIMIT $%d OFFSET $%d`,
		whereClause,
		mfilters.SortColumn(),
		mfilters.SortDirection(),
		argPosition,
		argPosition+1)

	args = append(args, mfilters.Limit(), mfilters.Offset())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := e.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	books := []*domain.Book{}

	for rows.Next() {
		var book domain.Book
		err := rows.Scan(
			&totalRecords,
			&book.ID,
			&book.Author,
			&book.Title,
			&book.MainGenre,
			&book.SubGenre,
			&book.Type,
			&book.Price,
			&book.Rating,
			&book.PeopleRated,
			&book.URL,
			&book.Version,
		)
		if err != nil {
			return nil, filters.Metadata{}, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, filters.Metadata{}, err
	}

	metadata := filters.CalculateMetadata(totalRecords, mfilters.Page, mfilters.PageSize)

	return books, metadata, nil
}

func (e BookModel) GetByGenre(genre string) ([]*domain.Book, error) {
	query := `
		SELECT id, author, title, main_genre, sub_genre, type, price, rating, people_rated, url, version
		FROM books
		WHERE main_genre = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := e.DB.QueryContext(ctx, query, genre)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*domain.Book

	for rows.Next() {
		var book domain.Book
		err := rows.Scan(
			&book.ID,
			&book.Author,
			&book.Title,
			&book.MainGenre,
			&book.SubGenre,
			&book.Type,
			&book.Price,
			&book.Rating,
			&book.PeopleRated,
			&book.URL,
			&book.Version,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func ValidateBook(v *validator.Validator, book *domain.Book) {
	v.Check(book.Author != "", "Author", "must be provided")
	v.Check(len(book.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(book.MainGenre != "", "main_genre", "must be provided")
	v.Check(book.SubGenre != "", "sub_genre", "must be provided")
	v.Check(book.Type != "", "type", "must be provided")
	v.Check(book.Price != "", "price", "must be provided")
	v.Check(book.Rating != 0, "rating", "must be provided")
	v.Check(book.PeopleRated != 0, "people_rated", "must be provided")
	v.Check(book.URL != "", "url", "must be provided")
}
