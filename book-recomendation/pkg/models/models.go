package models

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
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

type BookInteraction struct {
	UserID    int     `json:"user_id"`
	BookID    int     `json:"book_id"`
	Type      string  `json:"interaction_type"`
	Rating    float64 `json:"rating"`
	CreatedAt string  `json:"created_at"`
}

type RecommendationRequest struct {
	UserID         int      `json:"user_id"`
	UserBookTitles []string `json:"user_book_titles"`
}

type RecommendationResponse struct {
	RecommendedTitles []string `json:"recommended_titles"`
}

type UserPreferencesMap map[string]map[string]float64
