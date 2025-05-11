import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../auth.service';
import { User, AuthenticationResponse, ProfileUpdateData } from '../models';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css'],
})
export class ProfileComponent implements OnInit {
  userSession: AuthenticationResponse | null = null;
  isEditMode = false;
  updateData: ProfileUpdateData = {};
  updateMessage = '';
  updateError = '';
  showPassword = false;

  constructor(private authService: AuthService) {}

  ngOnInit(): void {
    this.loadUserProfile();
  }

  loadUserProfile(): void {
    const userData = sessionStorage.getItem('currentUser');
    if (userData) {
      try {
        this.userSession = JSON.parse(userData);
        // Initialize update data with current values
        if (this.userSession?.user) {
          this.updateData = {
            name: this.userSession.user.name,
            email: this.userSession.user.email,
            password: ''
          };
        }
      } catch (error) {
        console.error('Error parsing user data from session storage:', error);
      }
    }
  }

  toggleEditMode(): void {
    this.isEditMode = !this.isEditMode;
    this.updateMessage = '';
    this.updateError = '';
    
    // Reset form when canceling
    if (!this.isEditMode && this.userSession?.user) {
      this.updateData = {
        name: this.userSession.user.name,
        email: this.userSession.user.email,
        password: ''
      };
    }
  }

  saveProfile(): void {
    // Don't send password if it's empty
    const dataToUpdate: ProfileUpdateData = {
      name: this.updateData.name,
      email: this.updateData.email
    };
    
    if (this.updateData.password && this.updateData.password.trim() !== '') {
      dataToUpdate.password = this.updateData.password;
    }

    this.authService.updateUserProfile(dataToUpdate).subscribe({
      next: (response) => {
        this.userSession = response;
        this.updateMessage = 'Profile updated successfully!';
        this.updateError = '';
        this.isEditMode = false;
        
        // Reset password field
        this.updateData.password = '';
      },
      error: (error) => {
        console.error('Error updating profile:', error);
        this.updateError = error.error?.message || 'Failed to update profile. Please try again.';
        this.updateMessage = '';
      }
    });
  }

  togglePasswordVisibility(): void {
    this.showPassword = !this.showPassword;
  }

  formatDate(dateString: string): string {
    if (!dateString) return 'N/A';
    
    try {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch (error) {
      console.error('Error formatting date:', error);
      return dateString;
    }
  }
}