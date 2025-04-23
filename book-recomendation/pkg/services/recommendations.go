package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"book-recommendation-service/pkg/models"
)

type RecommendationService struct {
	client           *http.Client
	friendServiceURL string
}

func NewRecommendationService(friendServiceURL string) *RecommendationService {
	return &RecommendationService{
		client:           &http.Client{},
		friendServiceURL: friendServiceURL,
	}
}

func (s *RecommendationService) GetRecommendations(req models.RecommendationRequest) (*models.RecommendationResponse, error) {

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	url := s.friendServiceURL + "/recommendations"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from Python service: %s", resp.Status)
	}

	var response models.RecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &response, nil
}

func ConvertToPreferencesMap(prefs []models.UserPreference) models.UserPreferencesMap {
	result := make(models.UserPreferencesMap)

	for _, pref := range prefs {
		if _, exists := result[pref.Category]; !exists {
			result[pref.Category] = make(map[string]float64)
		}
		result[pref.Category][pref.Value] = pref.Weight
	}

	return result
}
