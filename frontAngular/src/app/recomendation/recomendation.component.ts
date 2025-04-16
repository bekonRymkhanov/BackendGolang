import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthenticationResponse, FavoriteBook } from '../models';
import { OneXBetService } from '../one-xbet.service';

interface RecomendationResponse {
  recommended_titles: string[];
}

@Component({
  selector: 'app-recomendation',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './recomendation.component.html',
  styleUrl: './recomendation.component.css'
})
export class RecomendationComponent implements OnInit {
  favoriteBooks: FavoriteBook[] = [];
  recommendedBooks: string[] = [];
  isLoading: boolean = false;
  errorMessage: string = '';
  userSession: AuthenticationResponse | null = null;
  showResults: boolean = false;
  
  constructor(private httpService: OneXBetService) {}
  
  ngOnInit(): void {
    this.loadUserProfile();
  }

  loadUserProfile(): void {
    const userData = sessionStorage.getItem('currentUser');
    if (userData) {
      try {
        this.userSession = JSON.parse(userData);
        this.loadFavoriteBooks();
      } catch (error) {
        console.error('Error parsing user data from session storage:', error);
        this.errorMessage = 'Error loading user profile. Please try logging in again.';
      }
    } else {
      this.errorMessage = 'Please log in to get personalized recommendations.';
    }
  }

  loadFavoriteBooks(): void {
    if (!this.userSession?.user) {
      this.errorMessage = 'User session not found. Please log in again.';
      return;
    }
    
    this.isLoading = true;
    this.httpService.getFavoriteBooks().subscribe(
      response => {
        this.favoriteBooks = response.favorite_books;
        this.isLoading = false;
      },
      error => {
        console.error('Error loading favorite books:', error);
        this.errorMessage = 'Error loading your favorite books. Please try again later.';
        this.isLoading = false;
      }
    );
  }

  getRecommendations(): void {
    if (!this.userSession?.user || this.favoriteBooks.length === 0) {
      this.errorMessage = this.favoriteBooks.length === 0 
        ? 'You need to add some books to your favorites first.' 
        : 'Please log in to get recommendations.';
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';
    this.showResults = false;
    
    // Extract book titles from favorite books
    const bookTitles = this.favoriteBooks.map(book => book.book_name);
    if (this.userSession.user.id === undefined) {
      this.errorMessage = 'User ID not found. Please log in again.';
      this.isLoading = false;
      return;
    }
    if (bookTitles.length === 0) {
      this.errorMessage = 'No favorite books found. Please add some books to your favorites.';
      this.isLoading = false;
      return;
    }
    // Make the recommendation request
    this.httpService.getBookRecommendations(this.userSession.user.id, bookTitles).subscribe(
      (response: RecomendationResponse) => {
        this.recommendedBooks = response.recommended_titles;
        this.showResults = true;
        this.isLoading = false;
      },
      error => {
        console.error('Error getting recommendations:', error);
        this.errorMessage = 'Failed to get recommendations. Please try again later.';
        this.isLoading = false;
      }
    );
  }
}
