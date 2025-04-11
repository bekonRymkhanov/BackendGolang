package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict raised")
)

type Models struct {
	Book     BookModel
	Comment  CommentModel
	Rating   RatingModel
	Genre    GenreModel
	SubGenre SubGenreModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Book:     BookModel{DB: db},
		Comment:  CommentModel{DB: db},
		Rating:   RatingModel{DB: db},
		Genre:    GenreModel{DB: db},
		SubGenre: SubGenreModel{DB: db},
	}
}
