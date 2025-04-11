package data

type SubGenre struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	MainGenre string  `json:"main_genre"`
	BookCount float64 `json:"book_count"`
	URL       string  `json:"url"`
	Version   int32   `json:"version"`
}
