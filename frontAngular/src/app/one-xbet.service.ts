import { Injectable } from '@angular/core';
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {User,Book,Genre,SubGenre,Comment,Rating,Metadata,BookFilters, requestBookDetail, FavoriteBook, AuthenticationResponse} from "./models";
import { Observable } from 'rxjs/internal/Observable';

@Injectable({
  providedIn: 'root'
})
export class OneXBetService {
  BACKEND_URL="http://localhost:4000"

  constructor(private client:HttpClient) { }

  private getAuthHeaders(): HttpHeaders {
    const userData = sessionStorage.getItem('currentUser');
    if (userData) {
      const user: { 
        token: string;
        expiry: string;
        user:User}= JSON.parse(userData);
      return new HttpHeaders({
        'Authorization': `Bearer ${user.token}`
      });
    }
    return new HttpHeaders();
  }
  getBooks(filters:BookFilters){
    return this.client.get<{ books: Book[], metadata: Metadata }>(`${this.BACKEND_URL}/Books/`, { params: { ...filters } })
  }
  getBook(id:number){
    return this.client.get<requestBookDetail>(`${this.BACKEND_URL}/Books/${id}/`)
  }
  deleteBook(id:number){
    return this.client.delete(`${this.BACKEND_URL}/Books/${id}/`)
  }
  postBook(newBook:Book){
    console.log(newBook.id)
    return this.client.post<Book>(`${this.BACKEND_URL}/Books/`,newBook)
  }
  putBook(newBook:Book){
    console.log(newBook.id)
    return this.client.put<Book>(`${this.BACKEND_URL}/Books/${newBook.id}/`,newBook)
  }
  getBookComments(id:number){
    console.log(`${this.BACKEND_URL}/booksComments/${id}/`)
    return this.client.get<{comments: Comment[], metadata: Metadata}>(`${this.BACKEND_URL}/booksComments/${id}/`)
  }
  getBookRatings(id:number){
    return this.client.get<Rating[]>(`${this.BACKEND_URL}/booksRatings/${id}/`)
  }



  getGenres(){
    return this.client.get<Genre[]>(`${this.BACKEND_URL}/Genres/`)
  }
  getGenre(id:number){
    return this.client.get<Genre>(`${this.BACKEND_URL}/Genres/${id}/`)
  }
  deleteGenre(id:number){
    return this.client.delete(`${this.BACKEND_URL}/Genre/${id}/`)
  }
  postGenre(newGenre:Genre){
    console.log(newGenre.id)
    return this.client.post<Genre>(`${this.BACKEND_URL}/Genre/`,newGenre)
  }
  putGenre(newGenre:Genre){
    console.log(newGenre.id)
    return this.client.put<Genre>(`${this.BACKEND_URL}/Genre/${newGenre.id}/`,newGenre)
  }
  getSubGenresByGenre(main_genre:string){
    return this.client.get<SubGenre[]>(`${this.BACKEND_URL}/Genre/${main_genre}/SubGenres`)
  }


  getSubGenres(){
    return this.client.get<SubGenre[]>(`${this.BACKEND_URL}/SubGenres/`)
  }
  getSubGenre(id:number){
    return this.client.get<SubGenre>(`${this.BACKEND_URL}/SubGenres/${id}/`)
  }
  deleteSubGenre(id:number){
    return this.client.delete(`${this.BACKEND_URL}/SubGenre/${id}/`)
  }
  postSubGenre(newSubGenre:SubGenre){
    console.log(newSubGenre.id)
    return this.client.post<SubGenre>(`${this.BACKEND_URL}/SubGenre/`,newSubGenre)
  }
  putSubGenre(newSubGenre:SubGenre){
    console.log(newSubGenre.id)
    return this.client.put<SubGenre>(`${this.BACKEND_URL}/SubGenre/${newSubGenre.id}/`,newSubGenre)
  }




  getComment(id:number){
    return this.client.get<Comment>(`${this.BACKEND_URL}/Comments/${id}/`)
  }
  deleteComment(id:number){
    return this.client.delete(`${this.BACKEND_URL}/Comments/${id}/`)
  }
  postComment(newComment:{
    book_id: number;
    user_id: number;
    content: string;
  }){
    return this.client.post<Comment>(`${this.BACKEND_URL}/Comments/`,newComment)
  }
  putComment(newComment:Comment){
    console.log(newComment.id)
    return this.client.put<Comment>(`${this.BACKEND_URL}/Comments/${newComment.id}/`,newComment)
  }



  getRating(id:number){
    return this.client.get<Rating>(`${this.BACKEND_URL}/Rating/${id}/`)
  }
  deleteRating(id:number){
    return this.client.delete(`${this.BACKEND_URL}/Rating/${id}/`)
  }
  postRating(newRating:Rating){
    console.log(newRating.id)
    return this.client.post<Rating>(`${this.BACKEND_URL}/Rating/`,newRating)
  }
  putRating(newRating:Rating){
    console.log(newRating.id)
    return this.client.put<Rating>(`${this.BACKEND_URL}/Rating/${newRating.id}/`,newRating)
  }
  getBookRatingByUser(bookID:number){
    return this.client.get<Rating>(`${this.BACKEND_URL}/Rating/${bookID}/`)
  }



  getFavoriteBooks() {
    const headers = this.getAuthHeaders();
    return this.client.get<{ favorite_books: FavoriteBook[] }>(`${this.BACKEND_URL}/favorite-books`,{headers});
  }
  
  addFavoriteBook(bookName: string) {
    const headers = this.getAuthHeaders();
    return this.client.post<{ favorite_book: FavoriteBook }>(`${this.BACKEND_URL}/favorite-books`, { book_name: bookName },{headers});
  }
  
  deleteFavoriteBook(id: number) {
    const headers = this.getAuthHeaders();
    return this.client.delete<{ message: string }>(`${this.BACKEND_URL}/favorite-books/${id}`,{headers});
  }


  getBookRecommendations(userId: number, bookTitles: string[]): Observable<any> {
    const headers = this.getAuthHeaders();
    // Make sure this URL points to your Go backend
    return this.client.post<any>('http://localhost:8080/recommendations', {
      user_id: userId,
      user_book_titles: bookTitles
    }, { headers });
  }



}


