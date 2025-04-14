import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject } from 'rxjs';
import { tap } from 'rxjs/operators';
import { User, AuthResponse, ActivationResponse, AuthenticationResponse } from './models';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = 'http://localhost:4000';
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  constructor(private http: HttpClient) {
    // Check session storage on service initialization
    const userData = sessionStorage.getItem('currentUser');
    if (userData) {
      this.currentUserSubject.next(JSON.parse(userData));
    }
  }

  register(user: User): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/users`, user).pipe(
      tap(response => {
        // Store registration token
        sessionStorage.setItem('registrationToken', response.token);
      })
    );
  }

  activateAccount(token: string): Observable<ActivationResponse> {
    return this.http.put<ActivationResponse>(`${this.apiUrl}/users/activated`, { "token":token });
  }

  login(email: string, password: string): Observable<AuthenticationResponse> {
    return this.http.post<AuthenticationResponse>(
      `${this.apiUrl}/tokens/authentication`, 
      { email, password }
    ).pipe(
      tap(response => {
        // After successful login, store the token and user info
        const userData = {
          token: response.authentication_token.token,
          expiry: response.authentication_token.expiry,
          user: response.user
        };
        sessionStorage.setItem('currentUser', JSON.stringify(userData));
        this.currentUserSubject.next(userData as unknown as User);
      })
    );
  }
  

  logout(): void {
    // Remove user data from session storage
    sessionStorage.removeItem('currentUser');
    this.currentUserSubject.next(null);
  }

  get isLoggedIn(): boolean {
    return !!sessionStorage.getItem('currentUser');
  }

  get currentUserValue(): User | null {
    return this.currentUserSubject.value;
  }

  get authToken(): string | null {
    const userData = sessionStorage.getItem('currentUser');
    if (userData) {
      const user = JSON.parse(userData);
      return user.token;
    }
    return null;
  }

  // Helper method to get auth headers for API calls that need authentication
  getAuthHeaders() {
    const token = this.authToken;
    if (token) {
      return { Authorization: `Bearer ${token}` };
    }
    return {};
  }
}