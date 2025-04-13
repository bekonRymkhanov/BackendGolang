package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/Books", app.listBooksHandler)
	router.HandlerFunc(http.MethodPost, "/Books", app.createBookHandler)
	router.HandlerFunc(http.MethodGet, "/Books/:id", app.showBookHandler)
	router.HandlerFunc(http.MethodPatch, "/Books/:id", app.updateBookHandler)
	router.HandlerFunc(http.MethodDelete, "/Books/:id", app.deleteBookHandler)

	router.HandlerFunc(http.MethodGet, "/Genres", app.listGenreHandler)
	router.HandlerFunc(http.MethodPost, "/Genres", app.createGenreHandler)
	router.HandlerFunc(http.MethodGet, "/Genres/:id", app.showGenreHandler)
	router.HandlerFunc(http.MethodPatch, "/Genres/:id", app.updateGenreHandler)
	router.HandlerFunc(http.MethodDelete, "/Genres/:id", app.deleteGenreHandler)

	router.HandlerFunc(http.MethodGet, "/SubGenres", app.listSubGenresHandler)
	router.HandlerFunc(http.MethodPost, "/SubGenres", app.createSubGenreHandler)
	router.HandlerFunc(http.MethodGet, "/SubGenres/:id", app.showSubGenreHandler)
	router.HandlerFunc(http.MethodPatch, "/SubGenres/:id", app.updateSubGenreHandler)
	router.HandlerFunc(http.MethodDelete, "/SubGenres/:id", app.deleteSubGenreHandler)
	router.HandlerFunc(http.MethodGet, "/Genre/:main_genre/SubGenres", app.showSubGenresByMainGenreHandler)

	router.HandlerFunc(http.MethodPost, "/comments", app.createCommentHandler)
	router.HandlerFunc(http.MethodGet, "/comments/:id", app.showCommentHandler)
	router.HandlerFunc(http.MethodPatch, "/comments/:id", app.updateCommentHandler)
	router.HandlerFunc(http.MethodDelete, "/comments/:id", app.deleteCommentHandler)
	router.HandlerFunc(http.MethodGet, "/booksComments/:id", app.listBookCommentsHandler)

	router.HandlerFunc(http.MethodPost, "/ratings", app.createRatingHandler)
	router.HandlerFunc(http.MethodGet, "/ratings/:id", app.showRatingHandler)
	router.HandlerFunc(http.MethodGet, "/booksRating/:bookID", app.showUserBookRatingHandler)
	router.HandlerFunc(http.MethodPatch, "/ratings/:id", app.updateRatingHandler)
	router.HandlerFunc(http.MethodDelete, "/ratings/:id", app.deleteRatingHandler)
	router.HandlerFunc(http.MethodGet, "/booksRatings/:id", app.listBookRatingsHandler)

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	return app.recoverPanic(app.rateLimit(router))

}
