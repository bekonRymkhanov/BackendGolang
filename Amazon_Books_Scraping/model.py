import pandas as pd
import numpy as np
from fuzzywuzzy import process
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.preprocessing import OneHotEncoder
from fastapi import FastAPI, HTTPException, Depends
import requests
from pydantic import BaseModel
from typing import List, Dict, Any
from collections import defaultdict

class ServiceAPI:
    def __init__(self, base_url: str):
        self.base_url = base_url 
    def get_user_preferences(self, user_id: int) -> Dict[str, Any]:
        response = requests.get(f"{self.base_url}/user/{user_id}/preferences")
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception("Ошибка получения предпочтений пользователя")
    def update_user_preferences(self, user_id: int, preferences: dict):
        response = requests.post(f"{self.base_url}/user/{user_id}/preferences", json=preferences)
        if response.status_code != 200:
            raise Exception("Ошибка обновления предпочтений пользователя")
    def get_global_preferences(self) -> Dict[str, Any]:
        response = requests.get(f"{self.base_url}/global/preferences")
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception("Ошибка получения глобальных предпочтений")

app = FastAPI()
class RecommendationRequest(BaseModel):
    user_id: int
    user_book_titles: List[str]
class RecommendationResponse(BaseModel):
    recommended_titles: List[str]

def get_db_service() -> ServiceAPI:
    return ServiceAPI(base_url="http://localhost:8000")

@app.post("/recommendations", response_model=RecommendationResponse)
def recommendations_endpoint(request: RecommendationRequest, db_api: ServiceAPI = Depends(get_db_service)):
    try:
        result = compute_recommendations(request.user_id, request.user_book_titles, db_api) # Лучше хранить тайтлы у юзеров
        return RecommendationResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

def collect_user_statistics(user_history, attributes):
    user_pref_sum = {attr: defaultdict(int) for attr in attributes}
    for book in user_history:
        for attr in attributes:
            value = book.get(attr)
            if value:
                user_pref_sum[attr][value] += 1
    user_count = len(user_history)
    return user_pref_sum, user_count

def bayesian_update(global_pref, user_pref_sum, user_count, prior_strength=10):
    updated_pref = {attr: {} for attr in global_pref}
    for attr in global_pref:
        for val in global_pref[attr]:
            prior = global_pref[attr][val]
            evidence = user_pref_sum[attr].get(val, 0)
            updated_pref[attr][val] = (prior_strength * prior + evidence) / (prior_strength + user_count)
    return updated_pref

def update_preferences_from_history(user_history, old_user_preferences, global_preferences, 
                                    attributes=['Main Genre', 'Sub Genre', 'Type', 'Author'], 
                                    prior_strength=10, alpha=0.1):
    user_pref_sum, user_count = collect_user_statistics(user_history, attributes)
    new_preferences = bayesian_update(global_preferences, user_pref_sum, user_count, prior_strength)

    updated_preferences = {attr: {} for attr in attributes}
    for attr in attributes:
        for val in global_preferences[attr]:
            old_value = old_user_preferences.get(attr, {}).get(val, global_preferences[attr].get(val, 0.5))
            new_value = new_preferences[attr].get(val, old_value)
            updated_preferences[attr][val] = (1 - alpha) * old_value + alpha * new_value
            
    return updated_preferences

def get_books_by_title(user_book_titles, clean_data):
    indices = []
    for title in user_book_titles:
        match = clean_data[clean_data['Title'].str.lower() == title.lower()]
        if not match.empty:
            indices.append(match.index[0])
        else:
            matches = process.extract(title, clean_data['Title'], limit=1)
            if matches and matches[0][1] > 60:
                closest_match = matches[0][0]
                fuzzy_match = clean_data[clean_data['Title'] == closest_match]
                if not fuzzy_match.empty:
                    indices.append(fuzzy_match.index[0])
                else:
                    print(f"Не найдено совпадение для книги (fuzzy): {title}")
            else:
                print(f"Не найдено совпадение для книги: {title}")
    return indices

def compute_explicit_weight(book, user_preferences):
    weight = 1.0
    for attr in ['Main Genre', 'Sub Genre', 'Type', 'Author']:
        attr_value = book[attr]
        if attr in user_preferences and attr_value in user_preferences[attr]:
            weight *= user_preferences[attr][attr_value]
    return weight

def zca_whitening(X, epsilon=1e-5):
    X_centered = X - np.mean(X, axis=0)
    sigma = np.cov(X_centered, rowvar=False)
    U, S, _ = np.linalg.svd(sigma)
    W = np.dot(U, np.dot(np.diag(1.0 / np.sqrt(S + epsilon)), U.T))
    X_whitened = np.dot(X_centered, W.T)
    return X_whitened



def compute_recommendations(user_id: int, user_book_titles: List[str], db_api: ServiceAPI) -> Dict[str, Any]:
    user_preferences = db_api.get_user_preferences(user_id)
    global_preferences = db_api.get_global_preferences()
    indices = get_books_by_title(user_book_titles, clean_data)
    user_books = [clean_data.iloc[ind] for ind in indices]
    user_preferences = update_preferences_from_history(user_books, user_preferences, global_preferences)
    db_api.update_user_preferences(user_id, user_preferences)
    book_preferences = [compute_explicit_weight(clean_data.iloc[idx], user_preferences) for idx in indices]
    books_vectors = features_whited[indices]
    user_profile = np.average(books_vectors, axis=0, weights=book_preferences)
    cosine_sim_scores = cosine_similarity(user_profile.reshape(1, -1), features_whited).flatten()
    top_indices = np.argsort(cosine_sim_scores)[::-1]
    filtered_indices = [idx for idx in top_indices if idx not in indices]
    rec_indices = filtered_indices[:10]
    recommended_titles = clean_data['Title'].iloc[rec_indices].tolist()
    return {"recommended_titles": recommended_titles}
db_api = ServiceAPI(base_url="http://backend-service.example.com/api")
books = pd.read_csv('/home/bekarys/Go/Amazon_Books_Scraping/Books_df.csv')  # change to your path
clean_data = books.reset_index(drop=True)
encoder = OneHotEncoder(sparse_output=False)
features = encoder.fit_transform(clean_data[['Main Genre', 'Sub Genre', 'Type', 'Author']])
features_whited = zca_whitening(features)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)    




