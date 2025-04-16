import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../auth.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule,RouterLink]
})
export class RegisterComponent implements OnInit {
  registerForm!: FormGroup;
  showActivation: boolean = false;
  activationToken: string = '';
  errorMessage: string = '';
  successMessage: string = '';
  loading: boolean = false;

  constructor(
    private formBuilder: FormBuilder,
    private authService: AuthService,
    private router: Router
  ) { }

  ngOnInit(): void {
    // Redirect if already logged in
    if (this.authService.isLoggedIn) {
      this.router.navigate(['/dashboard']);
    }

    this.registerForm = this.formBuilder.group({
      name: ['', [Validators.required]],
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(6)]],
      confirmPassword: ['', [Validators.required]]
    }, {
      validators: this.passwordMatchValidator
    });

    // Check if there's a registration token in session storage
    const token = sessionStorage.getItem('registrationToken');
    if (token) {
      this.activationToken = token;
      this.showActivation = true;
    }
  }

  passwordMatchValidator(formGroup: FormGroup) {
    const password = formGroup.get('password')?.value;
    const confirmPassword = formGroup.get('confirmPassword')?.value;
    
    if (password !== confirmPassword) {
      formGroup.get('confirmPassword')?.setErrors({ passwordMismatch: true });
      return { passwordMismatch: true };
    } else {
      return null;
    }
  }

  onSubmit(): void {
    // Stop here if form is invalid
    if (this.registerForm.invalid) {
      return;
    }

    this.loading = true;
    this.errorMessage = '';

    const user = {
      name: this.registerForm.value.name,
      email: this.registerForm.value.email,
      password: this.registerForm.value.password,
      is_admin: false
    };

    this.authService.register(user).subscribe({
      next: (response) => {
        this.activationToken = response.token;
        sessionStorage.setItem('registrationToken', response.token);
        this.showActivation = true;
        this.successMessage = 'Registration successful! Please activate your account.';
        this.loading = false;
      },
      error: (error) => {
        this.errorMessage = error.error?.message || 'Registration failed. Please try again.';
        this.loading = false;
      }
    });
  }

  activateAccount(): void {
    this.loading = true;
    this.errorMessage = '';
    console.log(this.activationToken);
    this.authService.activateAccount(this.activationToken).subscribe({
      next: () => {
        this.successMessage = 'Account activated successfully! You can now login.';
        sessionStorage.removeItem('registrationToken');
        this.loading = false;
        
        // Redirect to login page after a short delay
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 2000);
      },
      error: (error) => {
        this.errorMessage = error.error?.message || 'Account activation failed. Please try again.';
        this.loading = false;
      }
    });
  }
}
