package db

import (
	"database/sql"
	"fmt"
	"log"

	"book-recommendation-service/pkg/config"
	"book-recommendation-service/pkg/models"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(cfg config.DBConfig) (*Repository, error) {

	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Password,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("Successfully connected to the database")
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) GetUser(userID int) (*models.User, error) {
	query := `SELECT id, username, age, created_at FROM users WHERE id = $1`

	var user models.User
	err := r.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Age, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &user, nil
}

func (r *Repository) GetAllUsers() ([]models.User, error) {
	query := `SELECT id, username, age, created_at FROM users ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
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

func (r *Repository) GetUserPreferences(userID int) ([]models.UserPreference, error) {
	query := `SELECT user_id, category, value, weight, count 
              FROM user_preferences 
              WHERE user_id = $1
              ORDER BY category, weight DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var preferences []models.UserPreference
	for rows.Next() {
		var pref models.UserPreference
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

func (r *Repository) GetUserFavoriteBooks(userID int) ([]models.Book, error) {
	query := `SELECT b.id, b.title, b.author, b.main_genre, b.sub_genre, b.type
              FROM user_favorite_books uf
              JOIN books_amazon b ON uf.book_id = b.id
              WHERE uf.user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
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

func (r *Repository) GetUserBookInteractions(userID int) ([]models.BookInteraction, error) {
	query := `SELECT user_id, book_id, interaction_type, rating, created_at
              FROM user_book_interactions
              WHERE user_id = $1
              ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	var interactions []models.BookInteraction
	for rows.Next() {
		var interaction models.BookInteraction
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

func (r *Repository) GetGlobalPreferences() (models.UserPreferencesMap, error) {
	query := `SELECT category, value, avg_weight FROM global_preferences ORDER BY category, avg_weight DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	globalPrefs := models.UserPreferencesMap{
		"Main Genre": {},
		"Sub Genre":  {},
		"Type":       {},
		"Author":     {},
	}

	for rows.Next() {
		var category, value string
		var weight float64

		if err := rows.Scan(&category, &value, &weight); err != nil {
			return nil, fmt.Errorf("error scanning preference row: %v", err)
		}

		normalizedWeight := weight / 5.0

		if _, exists := globalPrefs[category]; exists {
			globalPrefs[category][value] = normalizedWeight
		} else {

			globalPrefs[category] = map[string]float64{value: normalizedWeight}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating preferences: %v", err)
	}

	return globalPrefs, nil
}

func (r *Repository) UpdateUserPreferences(userID int, prefs models.UserPreferencesMap) error {

	log.Printf("Updated preferences for user %d: %v", userID, prefs)

	return nil
}
