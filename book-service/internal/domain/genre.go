package domain

type Genre struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	SubgenreCount int    `json:"subgenre_count"`
	URL           string `json:"url"`
	Version       int32  `json:"version"`
}
