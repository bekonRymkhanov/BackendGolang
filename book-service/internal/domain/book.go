package domain

type Book struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	MainGenre   string  `json:"main_genre"`
	SubGenre    string  `json:"sub_genre"`
	Type        string  `json:"type"`
	Price       string  `json:"price"`
	Rating      float64 `json:"rating"`
	PeopleRated int64   `json:"people_rated"`
	URL         string  `json:"url"`
	Version     int32   `json:"version"`
}
