package main

import (
	"log"
	"net/http"

	"book-recommendation-service/pkg/api"
	"book-recommendation-service/pkg/config"
	"book-recommendation-service/pkg/db"
	"book-recommendation-service/pkg/services"
)

func main() {

	cfg := config.LoadConfig()

	repo, err := db.NewRepository(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	recService := services.NewRecommendationService(cfg.FriendServiceURL)

	handler := api.NewHandler(repo, recService)

	router := api.SetupRoutes(handler)

	log.Printf("Starting server on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
