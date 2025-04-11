// internal/service/auth_service.go
package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/pkg/crypto"
	"auth-service/pkg/validator"
	"context"
	"errors"
	"log"
	"time"
)

// TODO: нужно будет подраправить в будущем, что бы не висело здесь
var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrAccountDeactivated  = errors.New("account is deactivated")
	ErrUsernameExists      = errors.New("username already exists")
	ErrEmailExists         = errors.New("email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, input domain.UserRegisterInput) (*domain.User, error)
	Login(ctx context.Context, input domain.UserLoginInput) (*domain.Token, error)
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.Token, error)
	UpdateUser(ctx context.Context, userID uint, input domain.UserUpdateInput) (*domain.User, error)
	GetUsers(ctx context.Context, limit, offset int) ([]domain.User, int64, error)
	GetUserByID(ctx context.Context, id uint) (*domain.User, error)
	DeleteUser(ctx context.Context, id uint) error
	ChangeUserStatus(ctx context.Context, id uint, active bool) error
	ChangeUserRole(ctx context.Context, id uint, role string) error
}

type authService struct {
	userRepo    repository.UserRepository
	jwtService  JWTService
	tokenExpiry time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	jwtService JWTService,
	tokenExpiry time.Duration,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		jwtService:  jwtService,
		tokenExpiry: tokenExpiry,
	}
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, input domain.UserRegisterInput) (*domain.User, error) {

	v := validator.New()
	v.ValidateUsername("username", input.Username)
	v.ValidateEmail("email", input.Email)
	v.ValidatePassword("password", input.Password)
	v.ValidateName("first_name", input.FirstName)
	v.ValidateName("last_name", input.LastName)

	if !v.Valid() {
		return nil, errors.New("validation failed: " + formatValidationErrors(v.GetErrors()))
	}

	existingUser, err := s.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	existingUser, err = s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      "user", // Default role
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a token
func (s *authService) Login(ctx context.Context, input domain.UserLoginInput) (*domain.Token, error) {

	if input.Username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// Use same error for security reasons
		return nil, ErrInvalidCredentials
	}

	if !user.Active {
		return nil, ErrAccountDeactivated
	}

	if !crypto.CheckPasswordHash(input.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	user.LastLogin = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("Failed to update last login time: %v", err)
	}

	accessToken, err := s.jwtService.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
	}, nil
}

// ValidateToken validates a token and returns the associated user
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {

	if token == "" {
		return nil, ErrInvalidToken
	}

	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if !user.Active {
		return nil, ErrAccountDeactivated
	}

	return user, nil
}

// RefreshToken refreshes a token and returns a new token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.Token, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if !user.Active {
		return nil, ErrAccountDeactivated
	}

	accessToken, err := s.jwtService.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokenExpiry.Seconds()),
	}, nil
}

// UpdateUser updates user information
func (s *authService) UpdateUser(ctx context.Context, userID uint, input domain.UserUpdateInput) (*domain.User, error) {

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	changed := false

	if input.Email != nil && *input.Email != user.Email {
		v := validator.New()
		v.ValidateEmail("email", *input.Email)
		if !v.Valid() {
			return nil, errors.New("invalid email format")
		}

		existingUser, err := s.userRepo.FindByEmail(ctx, *input.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists
		}

		user.Email = *input.Email
		changed = true
	}

	if input.FirstName != nil {
		v := validator.New()
		v.ValidateName("first_name", *input.FirstName)
		if !v.Valid() {
			return nil, errors.New("invalid first name format")
		}

		user.FirstName = *input.FirstName
		changed = true
	}

	if input.LastName != nil {
		v := validator.New()
		v.ValidateName("last_name", *input.LastName)
		if !v.Valid() {
			return nil, errors.New("invalid last name format")
		}

		user.LastName = *input.LastName
		changed = true
	}

	if input.Password != nil && *input.Password != "" {
		v := validator.New()
		v.ValidatePassword("password", *input.Password)
		if !v.Valid() {
			return nil, errors.New("invalid password: " + formatValidationErrors(v.GetErrors()))
		}

		hashedPassword, err := crypto.HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
		changed = true
	}

	// обновлять только если были изминения
	if changed {
		user.UpdatedAt = time.Now()

		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

// GetUsers retrieves all users with pagination
func (s *authService) GetUsers(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {

	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.userRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserByID retrieves a user by ID
func (s *authService) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *authService) DeleteUser(ctx context.Context, id uint) error {

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}

// ChangeUserStatus activates or deactivates a user
func (s *authService) ChangeUserStatus(ctx context.Context, id uint, active bool) error {

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if user.Active != active {
		user.Active = active
		user.UpdatedAt = time.Now()

		return s.userRepo.Update(ctx, user)
	}

	return nil
}

// ChangeUserRole changes a user's role
func (s *authService) ChangeUserRole(ctx context.Context, id uint, role string) error {

	if role != "user" && role != "admin" {
		return errors.New("invalid role")
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if user.Role != role {
		user.Role = role
		user.UpdatedAt = time.Now()

		return s.userRepo.Update(ctx, user)
	}

	return nil
}

// Helper function to format validation errors
func formatValidationErrors(errors []validator.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	result := errors[0].Field + ": " + errors[0].Message
	if len(errors) > 1 {
		result += " and other errors"
	}

	return result
}
