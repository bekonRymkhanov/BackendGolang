package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/Books", app.listBooksHandler)                                               ///
	router.HandlerFunc(http.MethodPost, "/Books", app.requirePermission("movies:write", app.createBookHandler))      ////
	router.HandlerFunc(http.MethodGet, "/Books/:id", app.showBookHandler)                                            ////
	router.HandlerFunc(http.MethodPatch, "/Books/:id", app.requirePermission("movies:write", app.updateBookHandler)) ///
	router.HandlerFunc(http.MethodDelete, "/Books/:id", app.deleteBookHandler)                                       ////

	router.HandlerFunc(http.MethodGet, "/Genres", app.listGenreHandler)          ///
	router.HandlerFunc(http.MethodPost, "/Genres", app.createGenreHandler)       ///
	router.HandlerFunc(http.MethodGet, "/Genres/:id", app.showGenreHandler)      ////
	router.HandlerFunc(http.MethodPatch, "/Genres/:id", app.updateGenreHandler)  ////
	router.HandlerFunc(http.MethodDelete, "/Genres/:id", app.deleteGenreHandler) ///

	router.HandlerFunc(http.MethodGet, "/SubGenres", app.listSubGenresHandler)                              ////
	router.HandlerFunc(http.MethodPost, "/SubGenres", app.createSubGenreHandler)                            ////
	router.HandlerFunc(http.MethodGet, "/SubGenres/:id", app.showSubGenreHandler)                           ////
	router.HandlerFunc(http.MethodPatch, "/SubGenres/:id", app.updateSubGenreHandler)                       ///
	router.HandlerFunc(http.MethodDelete, "/SubGenres/:id", app.deleteSubGenreHandler)                      ////
	router.HandlerFunc(http.MethodGet, "/Genre/:main_genre/SubGenres", app.showSubGenresByMainGenreHandler) ///

	router.HandlerFunc(http.MethodPost, "/Comments", app.createCommentHandler)            ///
	router.HandlerFunc(http.MethodGet, "/Comments/:id", app.showCommentHandler)           ///
	router.HandlerFunc(http.MethodPatch, "/Comments/:id", app.updateCommentHandler)       ///
	router.HandlerFunc(http.MethodDelete, "/Comments/:id", app.deleteCommentHandler)      ///
	router.HandlerFunc(http.MethodGet, "/booksComments/:id", app.listBookCommentsHandler) ///

	router.HandlerFunc(http.MethodPost, "/Ratings", app.createRatingHandler)                  ////
	router.HandlerFunc(http.MethodGet, "/Ratings/:id", app.showRatingHandler)                 ///
	router.HandlerFunc(http.MethodGet, "/booksRating/:bookID", app.showUserBookRatingHandler) ///
	router.HandlerFunc(http.MethodPatch, "/Ratings/:id", app.updateRatingHandler)             ///
	router.HandlerFunc(http.MethodDelete, "/Ratings/:id", app.deleteRatingHandler)            ///
	router.HandlerFunc(http.MethodGet, "/booksRatings/:id", app.listBookRatingsHandler)       ///

	router.HandlerFunc(http.MethodPost, "/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodGet, "/favorite-books", app.requireAuthenticatedUser(app.GetFavoriteBooks))
	router.HandlerFunc(http.MethodPost, "/favorite-books", app.requireAuthenticatedUser(app.addFavoriteBookHandler))
	router.HandlerFunc(http.MethodDelete, "/favorite-books/:id", app.requireAuthenticatedUser(app.deleteFavoriteBookHandler))

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	return app.enableCORS(app.recoverPanic(app.rateLimit(app.authenticate(router))))

}
