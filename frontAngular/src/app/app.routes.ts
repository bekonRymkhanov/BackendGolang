import { Router, Routes } from '@angular/router';
import {HomeComponent} from "./home/home.component";
import {AboutComponent} from "./about/about.component";
import {NotFoundComponent} from "./not-found/not-found.component";
import { AuthService } from './auth.service';
import { inject } from '@angular/core';
import { AuthGuard } from './auth.guard';
import { BookDetailsComponent } from './book-details/book-details.component';
import { CommonModule } from '@angular/common';
import { BooksComponent } from './books/books.component';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { ProfileComponent } from './profile/profile.component';
import { FavoriteBooksComponent } from './favorite-books/favorite-books.component';
import { RecomendationComponent } from './recomendation/recomendation.component';



// Simple auth guard using functional guards (new in Angular)
export const authGuard = () => {
  const authService = inject(AuthService);
  const router = inject(Router);
  
  if (authService.isLoggedIn) {
    return true;
  }
  
  return router.parseUrl('/login');
};

export const routes: Routes = [
    { path:"",redirectTo:"home",pathMatch:"full" },
    { path:"home",component:HomeComponent,title:"Home page"},
    { path:"about",component:AboutComponent,title:"About page" },
    { path:"books",component:BooksComponent,title:"Books page" },
    { path:"books/:bookid",component:BookDetailsComponent,title:"Book details page" },
    { path: 'login', component: LoginComponent },
    { path: 'register', component: RegisterComponent },
    { path: 'profile', component: ProfileComponent },
    { path: 'favoriteBooks', component: FavoriteBooksComponent },
    { path: 'recomendation', component: RecomendationComponent },


    { path:"**",component:NotFoundComponent,title:"404 - not found" }
];
