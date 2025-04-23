package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"book-recommendation-service/pkg/db"
	"book-recommendation-service/pkg/models"
	"book-recommendation-service/pkg/services"
)

type Handler struct {
	repo       *db.Repository
	recService *services.RecommendationService
}

func NewHandler(repo *db.Repository, recService *services.RecommendationService) *Handler {
	return &Handler{
		repo:       repo,
		recService: recService,
	}
}

func (h *Handler) HandleRecommendationRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	recommendations, err := h.recService.GetRecommendations(req)
	if err != nil {
		log.Printf("Error getting recommendations: %v", err)
		http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

func (h *Handler) HandleUserPreferencesRequest(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	userIDStr := parts[2]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rawPrefs, err := h.repo.GetUserPreferences(userID)
	if err != nil {
		log.Printf("Error getting user preferences: %v", err)
		http.Error(w, "Failed to get user preferences", http.StatusInternalServerError)
		return
	}

	prefsMap := services.ConvertToPreferencesMap(rawPrefs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefsMap)
}

func (h *Handler) HandleUsersRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) <= 2 || parts[2] == "" {
		users, err := h.repo.GetAllUsers()
		if err != nil {
			log.Printf("Error getting users: %v", err)
			http.Error(w, "Failed to get users", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
		return
	}

	userIDStr := parts[2]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting user: %v", err)
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) HandleGlobalPreferencesRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	globalPrefs, err := h.repo.GetGlobalPreferences()
	if err != nil {
		log.Printf("Error fetching global preferences: %v", err)
		http.Error(w, "Failed to get global preferences", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globalPrefs)
}

func (h *Handler) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	userIDStr := parts[2]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var prefs models.UserPreferencesMap
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateUserPreferences(userID, prefs); err != nil {
		log.Printf("Error updating preferences: %v", err)
		http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}
