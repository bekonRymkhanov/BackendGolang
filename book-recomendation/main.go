package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Configuration for the service
type Config struct {
	FriendServiceURL string
	Port             string
}

// Database models based on actual schema
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"` // Omitted in responses for security
	Age       int    `json:"age"`
	CreatedAt string `json:"created_at"`
}

type UserPreference struct {
	UserID   int     `json:"user_id"`
	Category string  `json:"category"`
	Value    string  `json:"value"`
	Weight   float64 `json:"weight"`
	Count    int     `json:"count"`
}

type Book struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	MainGenre string `json:"main_genre"`
	SubGenre  string `json:"sub_genre"`
	Type      string `json:"type"`
}

// Request/Response structures for the recommendation API
type RecommendationRequest struct {
	UserID         int      `json:"user_id"`
	UserBookTitles []string `json:"user_book_titles"`
}

type RecommendationResponse struct {
	RecommendedTitles []string `json:"recommended_titles"`
}

// UserPreferencesMap for easier handling in the application
type UserPreferencesMap map[string]map[string]float64

// Create a new HTTP client
var client = &http.Client{}

var db *sql.DB

// setupDB establishes a connection to the PostgreSQL database
func setupDB() (*sql.DB, error) {
	// Get database connection params from environment
	host := getEnv("DB_HOST", "159.223.84.254")
	port := getEnv("DB_PORT", "5432")
	dbname := getEnv("DB_NAME", "maindb")
	user := getEnv("DB_USER", "dbadmin")
	password := getEnv("DB_PASSWORD", "cgroup123")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbname, user, password)

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	// Check connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("Successfully connected to the database")
	return db, nil
}
func setupCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin - you can restrict this to specific origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// getUserDetailsFromDB gets user information from the database
func getUserDetailsFromDB(userID int) (*User, error) {
	query := `SELECT id, username, age, created_at FROM users WHERE id = $1`

	var user User
	err := db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Age, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &user, nil
}

// getAllUsersFromDB gets all users from the database
func getAllUsersFromDB() ([]User, error) {
	query := `SELECT id, username, age, created_at FROM users ORDER BY id`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Age, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning user row: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}

	return users, nil
}

// getUserPreferencesFromDB gets raw user preferences from the database
func getUserPreferencesFromDB(userID int) ([]UserPreference, error) {
	query := `SELECT user_id, category, value, weight, count 
	          FROM user_preferences 
	          WHERE user_id = $1
	          ORDER BY category, weight DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var preferences []UserPreference
	for rows.Next() {
		var pref UserPreference
		if err := rows.Scan(&pref.UserID, &pref.Category, &pref.Value, &pref.Weight, &pref.Count); err != nil {
			return nil, fmt.Errorf("error scanning preference row: %v", err)
		}
		preferences = append(preferences, pref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating preferences: %v", err)
	}

	if len(preferences) == 0 {
		return nil, fmt.Errorf("no preferences found for user")
	}

	return preferences, nil
}

// getUserFavoriteBooksFromDB gets a user's favorite books from the database
func getUserFavoriteBooksFromDB(userID int) ([]Book, error) {
	query := `SELECT b.id, b.title, b.author, b.main_genre, b.sub_genre, b.type
	          FROM user_favorite_books uf
	          JOIN books_amazon b ON uf.book_id = b.id
	          WHERE uf.user_id = $1`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.MainGenre, &book.SubGenre, &book.Type); err != nil {
			return nil, fmt.Errorf("error scanning book row: %v", err)
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating books: %v", err)
	}

	return books, nil
}

// getUserBookInteractionsFromDB gets a user's book interactions from the database
func getUserBookInteractionsFromDB(userID int) ([]BookInteraction, error) {
	query := `SELECT user_id, book_id, interaction_type, rating, created_at
	          FROM user_book_interactions
	          WHERE user_id = $1
	          ORDER BY created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var interactions []BookInteraction
	for rows.Next() {
		var interaction BookInteraction
		if err := rows.Scan(&interaction.UserID, &interaction.BookID, &interaction.Type, &interaction.Rating, &interaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning interaction row: %v", err)
		}
		interactions = append(interactions, interaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating interactions: %v", err)
	}

	return interactions, nil
}

// Additional type needed for book interactions
type BookInteraction struct {
	UserID    int     `json:"user_id"`
	BookID    int     `json:"book_id"`
	Type      string  `json:"interaction_type"`
	Rating    float64 `json:"rating"`
	CreatedAt string  `json:"created_at"`
}

func getRecommendations(config Config, req RecommendationRequest) (*RecommendationResponse, error) {
	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	// Create request to Python service
	url := config.FriendServiceURL + "/recommendations"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from Python service: %s", resp.Status)
	}

	// Parse response
	var response RecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &response, nil
}

// getUserRawPreferences gets raw user preferences from the database
func getUserRawPreferences(userID int) ([]UserPreference, error) {
	return getUserPreferencesFromDB(userID)
}

// convertToPreferencesMap converts raw preference data to the map format
func convertToPreferencesMap(prefs []UserPreference) UserPreferencesMap {
	result := make(UserPreferencesMap)

	for _, pref := range prefs {
		if _, exists := result[pref.Category]; !exists {
			result[pref.Category] = make(map[string]float64)
		}
		result[pref.Category][pref.Value] = pref.Weight
	}

	return result
}

// handleRecommendationRequest handles incoming recommendation requests
func handleRecommendationRequest(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req RecommendationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Get recommendations from Python service
		recommendations, err := getRecommendations(config, req)
		if err != nil {
			log.Printf("Error getting recommendations: %v", err)
			http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recommendations)
	}
}

// handleUserPreferencesRequest handles requests for user preferences
func handleUserPreferencesRequest(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from URL path
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
		rawPrefs, dbErr := getUserRawPreferences(userID)
		if dbErr != nil {
			log.Printf("Error getting user preferences: %v", dbErr)
			http.Error(w, "Failed to get user preferences", http.StatusInternalServerError)
			return
		}
		fmt.Println(rawPrefs)

		// Convert to map format
		prefsMap := convertToPreferencesMap(rawPrefs)
		prefs := &prefsMap

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
	}
}

// handleUsersRequest handles requests for user information
func handleUsersRequest() http.HandlerFunc {
	fmt.Println("Handling user requests")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from path if specified
		parts := strings.Split(r.URL.Path, "/")

		// List all users
		if len(parts) <= 2 || parts[2] == "" {
			users, err := getAllUsersFromDB()
			if err != nil {
				log.Printf("Error getting users: %v", err)
				http.Error(w, "Failed to get users", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
			return
		}

		// Get specific user
		userIDStr := parts[2]
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		user, err := getUserDetailsFromDB(userID)
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
}

func handleGlobalPreferencesRequest() http.HandlerFunc {
	fmt.Println("Handling global preferences requests")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Query global preferences from database
		query := `SELECT category, value, avg_weight FROM global_preferences ORDER BY category, avg_weight DESC`

		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Error fetching global preferences: %v", err)
			http.Error(w, "Failed to get global preferences", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Prepare the response structure
		globalPrefs := map[string]map[string]float64{
			"Main Genre": {},
			"Sub Genre":  {},
			"Type":       {},
			"Author":     {},
		}

		// Populate from database results
		for rows.Next() {
			var category, value string
			var weight float64

			if err := rows.Scan(&category, &value, &weight); err != nil {
				log.Printf("Error scanning preference row: %v", err)
				continue
			}

			// Normalize weight to 0-1 range (assuming weights in DB are on a scale like 1-5)
			normalizedWeight := weight / 5.0

			// Add to appropriate category
			if _, exists := globalPrefs[category]; exists {
				globalPrefs[category][value] = normalizedWeight
			} else {
				// Handle unexpected category
				globalPrefs[category] = map[string]float64{value: normalizedWeight}
			}
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating preferences: %v", err)
			http.Error(w, "Error processing preferences", http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globalPrefs)
	}
}

func handleUpdateUserPreferences() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from URL path
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

		// Parse request body into preference map
		var prefs map[string]map[string]float64
		if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// In a real system, you would save these preferences to your database
		// For now, we'll just respond with OK
		log.Printf("Updated preferences for user %d: %v", userID, prefs)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	}
}

func main() {
	// Load configuration
	config := Config{
		FriendServiceURL: getEnv("FRIEND_SERVICE_URL", "http://localhost:8001"),
		Port:             getEnv("PORT", "8080"),
	}
	var err error
	db, err = setupDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a new router
	mux := http.NewServeMux()

	// Set up routes
	mux.HandleFunc("/recommendations", handleRecommendationRequest(config))
	mux.HandleFunc("/global/preferences", handleGlobalPreferencesRequest())

	mux.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		// Split the path to determine if this is a preferences request
		parts := strings.Split(r.URL.Path, "/")

		// Check if this is a preferences request
		if len(parts) >= 4 && parts[3] == "preferences" {
			if r.Method == http.MethodGet {
				handleUserPreferencesRequest(config)(w, r)
			} else if r.Method == http.MethodPost {
				handleUpdateUserPreferences()(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Otherwise handle as a regular user request
		http.NotFound(w, r)
	})
	mux.HandleFunc("/users/", handleUsersRequest())

	// Serve static files for frontend
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// Apply CORS middleware to all routes
	handler := setupCORS(mux)

	// Start server
	log.Printf("Starting server on port %s", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, handler))
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
