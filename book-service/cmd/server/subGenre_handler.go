package main

import (
	"book-service/internal/data"
	"book-service/internal/domain"
	"book-service/internal/filters"
	"book-service/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createSubGenreHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title     string  `json:"title"`
		MainGenre string  `json:"main_genre"`
		BookCount float64 `json:"book_count"`
		URL       string  `json:"url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	sgenre := &domain.SubGenre{
		Title:     input.Title,
		MainGenre: input.MainGenre,
		BookCount: input.BookCount,
		URL:       input.URL,
	}

	if data.ValidateSubGenre(v, sgenre); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.SubGenre.Insert(sgenre)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/SubGenres/%d", sgenre.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"sub genre": sgenre}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showSubGenreHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	subGenre, err := app.models.SubGenre.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sub_genre": subGenre}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateSubGenreHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	subGenre, err := app.models.SubGenre.Get(id)
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
		Title     *string  `json:"title"`
		MainGenre *string  `json:"main_genre"`
		BookCount *float64 `json:"book_count"`
		URL       *string  `json:"url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		subGenre.Title = *input.Title
	}
	if input.MainGenre != nil {
		subGenre.MainGenre = *input.MainGenre
	}
	if input.BookCount != nil {
		subGenre.BookCount = *input.BookCount
	}
	if input.URL != nil {
		subGenre.URL = *input.URL
	}

	v := validator.New()

	if data.ValidateSubGenre(v, subGenre); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.SubGenre.Update(subGenre)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sub_genre": subGenre}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteSubGenreHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.SubGenre.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "sub genre successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSubGenresHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		filters.Filters
	}
	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafelist = []string{
		"id", "title", "main_genre", "book_count", "-id", "-title", "-main_genre", "-book_count",
	}

	if filters.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	subGenres, metadata, err := app.models.SubGenre.GetAll(input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"sub_genres": subGenres, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showSubGenresByMainGenreHandler(w http.ResponseWriter, r *http.Request) {
	mainGenre := r.URL.Query().Get("main_genre")

	subGenres, err := app.models.SubGenre.GetByGenre(mainGenre)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sub_genres": subGenres}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
