import {Component, OnInit} from '@angular/core';
import {RouterLink, RouterOutlet} from '@angular/router';
import { CommonModule } from '@angular/common';
import {UserComponent} from "./user/user.component";
import {FormsModule} from "@angular/forms";
import {OneXBetService} from "./one-xbet.service";
import {HomeComponent} from "./home/home.component";
import {AboutComponent} from "./about/about.component";
import {BooksComponent} from "./books/books.component";
import { BookDetailsComponent } from './book-details/book-details.component';
import { NotFoundComponent } from './not-found/not-found.component';
import { Router } from '@angular/router';
import { AuthService } from './auth.service';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { ProfileComponent } from './profile/profile.component';
@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet,BooksComponent,ProfileComponent,LoginComponent,NotFoundComponent,HomeComponent,RegisterComponent,BookDetailsComponent, CommonModule, UserComponent, RouterLink, FormsModule,HomeComponent,AboutComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent implements OnInit{
  title= 'Book Recomendation System';
  isLoggedIn:boolean=false;
  constructor(private httpService:OneXBetService,private authService:AuthService) {
    this.authService=authService;
  }

  ngOnInit(): void {
    this.isLoggedIn=this.authService.isLoggedIn;
  }

  protected readonly localStorage = localStorage;
  logout() {
    this.authService.logout();
    this.isLoggedIn = false;
  }
}
