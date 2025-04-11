package repository

import (
	"auth-service/internal/domain"
	"context"
)

type UserRepository interface {
	// Create adds a new user to the database
	Create(ctx context.Context, user *domain.User) error

	// FindByUsername finds a user by username
	FindByUsername(ctx context.Context, username string) (*domain.User, error)

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id uint) (*domain.User, error)

	// FindAll retrieves all users with pagination
	FindAll(ctx context.Context, limit, offset int) ([]domain.User, error)

	// Update updates user information
	Update(ctx context.Context, user *domain.User) error

	// Delete removes a user from the database
	Delete(ctx context.Context, id uint) error

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// Search users by username, email, or name
	Search(ctx context.Context, query string, limit, offset int) ([]domain.User, error)
}
