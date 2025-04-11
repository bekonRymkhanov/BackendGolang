package main

import (
	"book-service/internal/data"
	"book-service/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string  `json:"title"`
		Author      string  `json:"author"`
		MainGenre   string  `json:"main_genre"`
		SubGenre    string  `json:"sub_genre"`
		Type        string  `json:"type"`
		Price       string  `json:"price"`
		Rating      float64 `json:"rating"`
		PeopleRated int64   `json:"people_rated"`
		URL         string  `json:"url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	book := &data.Book{
		Title:       input.Title,
		Author:      input.Author,
		MainGenre:   input.MainGenre,
		SubGenre:    input.SubGenre,
		Type:        input.Type,
		Price:       input.Price,
		Rating:      input.Rating,
		PeopleRated: input.PeopleRated,
		URL:         input.URL,
	}

	if data.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Book.Insert(book)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/books/%d", book.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"book": book}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	book, err := app.models.Book.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	book, err := app.models.Book.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title       *string  `json:"title"`
		Author      *string  `json:"author"`
		MainGenre   *string  `json:"main_genre"`
		SubGenre    *string  `json:"sub_genre"`
		Type        *string  `json:"type"`
		Price       *string  `json:"price"`
		Rating      *float64 `json:"rating"`
		PeopleRated *int64   `json:"people_rated"`
		URL         *string  `json:"url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Author != nil {
		book.Author = *input.Author
	}
	if input.MainGenre != nil {
		book.MainGenre = *input.MainGenre
	}
	if input.PeopleRated != nil {
		book.PeopleRated = *input.PeopleRated
	}
	if input.Price != nil {
		book.Price = *input.Price
	}
	if input.Rating != nil {
		book.Rating = *input.Rating
	}
	if input.SubGenre != nil {
		book.SubGenre = *input.SubGenre
	}
	if input.Title != nil {
		book.Title = *input.Title
	}
	if input.Type != nil {
		book.Type = *input.Type
	}
	if input.URL != nil {
		book.URL = *input.URL
	}

	v := validator.New()

	if data.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Book.Update(book)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Book.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "book successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		data.Filters
	}
	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafelist = []string{"id", "title", "author", "main_genre", "sub_genre", "type", "price", "rating", "people_rated", "-id", "-title", "-author", "-main_genre", "-sub_genre", "-type", "-price", "-rating", "-people_rated"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, metadata, err := app.models.Book.GetAll(input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"books": books, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) showBooksByGenreHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the genre from the URL parameters
	genre := r.URL.Query().Get("genre")

	books, err := app.models.Book.GetByGenre(genre)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the characters as a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"books": books}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
