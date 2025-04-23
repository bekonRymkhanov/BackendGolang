package api

import (
	"net/http"
	"strings"

	"book-recommendation-service/pkg/middleware"
)

func SetupRoutes(handler *Handler) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/recommendations", handler.HandleRecommendationRequest)
	mux.HandleFunc("/global/preferences", handler.HandleGlobalPreferencesRequest)

	mux.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {

		parts := strings.Split(r.URL.Path, "/")

		if len(parts) >= 4 && parts[3] == "preferences" {
			if r.Method == http.MethodGet {
				handler.HandleUserPreferencesRequest(w, r)
			} else if r.Method == http.MethodPost {
				handler.HandleUpdateUserPreferences(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("/users/", handler.HandleUsersRequest)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	return middleware.CORS(mux)
}
