import { Component } from '@angular/core';
import { AuthenticationResponse, Book, FavoriteBook } from '../models';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { AuthService } from '../auth.service';
import { OneXBetService } from '../one-xbet.service';
import { NgIf, NgForOf } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-favorite-books',
  standalone: true,
  imports: [    NgIf,
    FormsModule,
    RouterLink,
    NgForOf],
  templateUrl: './favorite-books.component.html',
  styleUrl: './favorite-books.component.css'
})

export class FavoriteBooksComponent {
  book!: Book;
  favoriteBooks: FavoriteBook[] = [];
  loaded: boolean = false;
  removingFromFavorites: boolean = false;


  userSession: AuthenticationResponse | null = null;

  constructor(
    private httpService: OneXBetService,
    private authService: AuthService,
    private route: ActivatedRoute
  ) {

  }
  
  ngOnInit(): void {
    this.loaded = false;
    this.loadUserProfile();

  }
  loadUserProfile(): void {
    const userData = sessionStorage.getItem('currentUser');
    if (userData) {
      try {
        this.userSession = JSON.parse(userData);
        // Load favorite books after user is authenticated
        this.loadFavoriteBooks();
      } catch (error) {
        console.error('Error parsing user data from session storage:', error);
      }
    }
    console.log(this.favoriteBooks)
    this,this.loaded = true;

  }
  loadFavoriteBooks(): void {
    if (!this.userSession?.user) {
      return;
    }
    
    this.httpService.getFavoriteBooks().subscribe(
      response => {
        this.favoriteBooks = response.favorite_books;
      },
      error => {
        console.error('Error loading favorite books:', error);
      }
    );
  }
  removeFromFavorites(favorite_book_id: number): void {
    if (!this.userSession?.user) {
      return;
    }
    
    this.removingFromFavorites = true;
    this.httpService.deleteFavoriteBook(favorite_book_id).subscribe(
      response => {
        this.favoriteBooks = this.favoriteBooks.filter(favorite_book => favorite_book_id !== favorite_book.id);
        this.removingFromFavorites = false;
      },
      error => {
        console.error('Error removing from favorites:', error);
        this.removingFromFavorites = false;
      }
    );

  }


}
