import { Component, OnInit } from '@angular/core';
import { AuthenticationResponse, Book, Comment, FavoriteBook } from '../models';
import { AuthService } from '../auth.service';
import { OneXBetService } from '../one-xbet.service';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { NgIf, NgForOf } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-book-details',
  standalone: true,
  imports: [
    NgIf,
    FormsModule,
    RouterLink,
    NgForOf
  ],
  templateUrl: './book-details.component.html',
  styleUrl: './book-details.component.css'
})
export class BookDetailsComponent implements OnInit {
  book!: Book;
  loaded: boolean = false;
  comments: Comment[] = [];
  favoriteBooks: FavoriteBook[] = [];
  currentFavorite: FavoriteBook | null = null;
  isFavorite: boolean = false;
  addingToFavorites: boolean = false;
  removingFromFavorites: boolean = false;
  
  newComment = {
    "book_id": 0,
    "content": '',
    "user_id": 0,
  };
  userSession: AuthenticationResponse | null = null;

  constructor(
    private httpService: OneXBetService,
    private authService: AuthService,
    private route: ActivatedRoute,
    private router: Router
  ) {
    this.newComment = {
      "book_id": 0,
      "content": '',
      "user_id": 0,
    }
  }
  
  ngOnInit(): void {
    this.loaded = false;
    this.loadUserProfile();
    
    this.route.params.subscribe(params => {
      const bookId = params['bookid'];
      this.httpService.getBook(bookId).subscribe(request => {
        this.book = request.book;
        // After loading the book, check if it's in favorites
        this.checkIfBookIsFavorite();
      });

      this.httpService.getBookComments(bookId).subscribe(response => {
        this.comments = response.comments;
        this.loaded = true;
      });
    });
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
  }

  loadFavoriteBooks(): void {
    if (!this.userSession?.user) {
      return;
    }
    
    this.httpService.getFavoriteBooks().subscribe(
      response => {
        this.favoriteBooks = response.favorite_books;
        this.checkIfBookIsFavorite();
      },
      error => {
        console.error('Error loading favorite books:', error);
      }
    );
  }

  checkIfBookIsFavorite(): void {
    if (!this.book || !this.favoriteBooks.length) {
      this.isFavorite = false;
      this.currentFavorite = null;
      return;
    }

    const foundFavorite = this.favoriteBooks.find(fb => fb.book_name === this.book.title);
    this.isFavorite = !!foundFavorite;
    this.currentFavorite = foundFavorite || null;
  }

  toggleFavorite(): void {
    if (!this.userSession?.user) {
      alert('Please log in to add favorites');
      return;
    }

    if (this.isFavorite && this.currentFavorite) {
      this.removeFromFavorites();
    } else {
      this.addToFavorites();
    }
  }

  addToFavorites(): void {
    if (!this.book || this.addingToFavorites) {
      return;
    }

    this.addingToFavorites = true;
    this.httpService.addFavoriteBook(this.book.title).subscribe(
      response => {
        this.favoriteBooks.push(response.favorite_book);
        this.currentFavorite = response.favorite_book;
        this.isFavorite = true;
        this.addingToFavorites = false;
      },
      error => {
        console.error('Error adding book to favorites:', error);
        this.addingToFavorites = false;
        if (error.status === 409) {
          alert('This book is already in your favorites');
        }
      }
    );
  }

  removeFromFavorites(): void {
    if (!this.currentFavorite || this.removingFromFavorites) {
      return;
    }

    this.removingFromFavorites = true;
    this.httpService.deleteFavoriteBook(this.currentFavorite.id).subscribe(
      response => {
        this.favoriteBooks = this.favoriteBooks.filter(fb => fb.id !== this.currentFavorite?.id);
        this.isFavorite = false;
        this.currentFavorite = null;
        this.removingFromFavorites = false;
      },
      error => {
        console.error('Error removing book from favorites:', error);
        this.removingFromFavorites = false;
      }
    );
  }

  // Existing methods
  EditComment(comment: Comment) {
    this.httpService.putComment(comment).subscribe(response => {
      if (response && response) {
        const index = this.comments.findIndex(c => c.id === comment.id);
        if (index !== -1) {
          this.comments[index] = response;
        }
      }
    }, error => {
      console.error('Error updating comment:', error);
    });
  }

  DeleteComment(id: number) {
    this.httpService.deleteComment(id).subscribe(() => {
      this.comments = this.comments.filter(comment => comment.id !== id);
    }, error => {
      console.error('Error deleting comment:', error);
    });
  }

  CreateComments() {
    if (this.book && this.userSession?.user) {
      this.newComment.book_id = this.book.id;
      this.newComment.user_id = this.userSession.user.id ?? 0;
      
      this.httpService.postComment(this.newComment).subscribe(response => {
        if (response) {
          this.comments.unshift(response);       
          this.newComment = {
            "book_id": this.book.id,
            "content": '',
            "user_id": this.userSession?.user?.id ?? 0,
          };
        }
      });
    }
  }
  deleteBook() {
    if (this.book) {
      this.httpService.deleteBook(this.book.id).subscribe(() => {
        alert('Book deleted successfully');
        this.router.navigate(['/books']);
      }, error => {
        console.error('Error deleting book:', error);
      });
    }
  }
}