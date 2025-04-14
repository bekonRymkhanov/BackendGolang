package main

import (
	"book-service/internal/domain"
	"book-service/internal/validator"
	"errors"
	"fmt"
	"net/http"

	"book-service/internal/data"
)

// Add to Models struct in data.go file:
// FavoriteBook FavoriteBookModel

func (app *application) GetFavoriteBooks(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the authenticated user
	user := app.contextGetUser(r)
	if user.ID == 0 {
		app.authenticationRequiredResponse(w, r)
		return
	}

	// Get all favorite books for the user
	favoriteBooks, err := app.models.FavoriteBook.GetAllForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"favorite_books": favoriteBooks}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) addFavoriteBookHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the authenticated user
	user := app.contextGetUser(r)
	if user.ID == 0 {
		app.authenticationRequiredResponse(w, r)
		return
	}

	// Parse and validate the request
	var input struct {
		BookName string `json:"book_name"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	favoriteBook := &domain.FavoriteBook{
		UserID:   user.ID,
		BookName: input.BookName,
	}

	v := validator.New()
	data.ValidateFavoriteBook(v, favoriteBook)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the favorite book
	err = app.models.FavoriteBook.Insert(favoriteBook)
	if err != nil {
		// Check for duplicate (could be more specific with a custom PostgreSQL error check)
		if app.isDuplicateErr(err) {
			app.writeJSON(w, http.StatusConflict, envelope{"error": "This book is already in your favorites"}, nil)
			return
		}

		app.serverErrorResponse(w, r, err)
		return
	}

	// Return the created favorite book
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/favorite-books/%d", favoriteBook.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"favorite_book": favoriteBook}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFavoriteBookHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the book ID from the URL
	id, err := app.readIDParam(w, r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Extract the user ID from the authenticated user
	user := app.contextGetUser(r)
	if user.ID == 0 {
		app.authenticationRequiredResponse(w, r)
		return
	}

	// Delete the favorite book
	err = app.models.FavoriteBook.Delete(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "favorite book successfully removed"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Helper method to check for duplicate errors (you may need to adjust this based on your error handling)
func (app *application) isDuplicateErr(err error) bool {
	// This is a simplified check. In real world, check for specific PostgreSQL error code
	return err != nil && err.Error() != "" && contains(err.Error(), "duplicate") && contains(err.Error(), "unique")
}

func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) >= len(substr) && s != substr && fmt.Sprintf("%s", s) != fmt.Sprintf("%s", substr) && s != fmt.Sprintf("%s", substr) && fmt.Sprintf("%s", s) != substr
}
