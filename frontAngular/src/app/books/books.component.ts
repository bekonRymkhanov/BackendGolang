import { Component } from '@angular/core';
import {NgForOf, NgIf} from "@angular/common";
import {FormsModule} from "@angular/forms";
import {RouterLink} from "@angular/router";
import {OneXBetService} from "../one-xbet.service";
import {Book, Metadata,BookFilters} from "../models";
@Component({
  selector: 'app-books',
  standalone: true,
  imports: [
    NgIf,
    FormsModule,
    RouterLink,
    NgForOf
  ],
  templateUrl: './books.component.html',
  styleUrl: './books.component.css'
})

export class BooksComponent {
  books: Book[] = [];
  metadata!: Metadata;
  loaded = false;

  filters: BookFilters = {
    title: '',
    author: '',
    main_genre: ''
  };

  constructor(private httpService: OneXBetService) {}

  ngOnInit(): void {
    this.loadBooks();
  }

  loadBooks(): void {
    this.loaded = false;
    this.httpService.getBooks(this.filters).subscribe(response => {
      this.books = response.books;
      this.metadata = response.metadata;
      this.loaded = true;
    });

  }

  onSearch(): void {
    this.filters = { ...this.filters, page: 1 };
    
    this.loadBooks();
  }

  onClear(): void {
    this.filters = { title: '', author: '', main_genre: '' };
    this.loadBooks();
  }
  NextPage(): void {
    if (this.metadata.CurrentPage >= this.metadata.LastPage) {
      return;
    }
    this.filters = { ...this.filters, page: this.metadata.CurrentPage + 1 };
    this.loadBooks();
  }
  PreviousPage(): void {
    if (this.metadata.CurrentPage <= 1) {
      return;
    }
    this.filters = { ...this.filters, page: this.metadata.CurrentPage - 1 };
    this.loadBooks();
  }

}
