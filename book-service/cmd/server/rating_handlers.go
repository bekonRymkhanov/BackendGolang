package main

import (
	"book-service/internal/data"
	"book-service/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createRatingHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BookID int64 `json:"book_id"`
		Score  int   `json:"score"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userID := app.getUserIDFromRequest(r) // tipo beret userid otkuda to(posmotri funcciu)

	v := validator.New()

	rating := &data.Rating{
		BookID: input.BookID,
		UserID: userID,
		Score:  input.Score,
	}

	if data.ValidateRating(v, rating); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	_, err = app.models.Book.Get(input.BookID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("book_id", "book does not exist")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Rating.Insert(rating)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	avgRating, count, err := app.models.Rating.GetAverageRating(input.BookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	book, err := app.models.Book.Get(input.BookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	book.Rating = avgRating
	book.PeopleRated = int64(count)

	err = app.models.Book.Update(book)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/ratings/%d", rating.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"rating": rating}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showRatingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	rating, err := app.models.Rating.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"rating": rating}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showUserBookRatingHandler(w http.ResponseWriter, r *http.Request) {

	bookID, err := app.readBookIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID := app.getUserIDFromRequest(r)

	rating, err := app.models.Rating.GetUserRatingForBook(userID, bookID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"rating": rating}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateRatingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID := app.getUserIDFromRequest(r) // userid nado

	rating, err := app.models.Rating.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if rating.UserID != userID {
		app.notPermittedResponse(w, r)
		return
	}

	var input struct {
		Score *int `json:"score"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Score != nil {
		rating.Score = *input.Score
	}

	v := validator.New()

	if data.ValidateRating(v, rating); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Rating.Update(rating)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	avgRating, count, err := app.models.Rating.GetAverageRating(rating.BookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	book, err := app.models.Book.Get(rating.BookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	book.Rating = avgRating
	book.PeopleRated = int64(count)

	err = app.models.Book.Update(book)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"rating": rating}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteRatingHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID := app.getUserIDFromRequest(r) //userid nuzhen

	rating, err := app.models.Rating.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	bookID := rating.BookID

	err = app.models.Rating.Delete(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	avgRating, count, err := app.models.Rating.GetAverageRating(bookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	book, err := app.models.Book.Get(bookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	book.Rating = avgRating
	book.PeopleRated = int64(count)

	err = app.models.Book.Update(book)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "rating successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBookRatingsHandler(w http.ResponseWriter, r *http.Request) {
	bookID, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	_, err = app.models.Book.Get(bookID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	ratings, err := app.models.Rating.GetAllForBook(bookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	avgRating, count, err := app.models.Rating.GetAverageRating(bookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	response := envelope{
		"ratings": ratings,
		"summary": map[string]interface{}{
			"average": avgRating,
			"count":   count,
		},
	}

	err = app.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
