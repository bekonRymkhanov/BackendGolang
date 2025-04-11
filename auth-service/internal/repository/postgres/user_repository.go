package postgres

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding user by username: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding user by ID: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	query := r.db.WithContext(ctx).Model(&domain.User{})

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	query = query.Order("id ASC")

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("error finding all users: %w", err)
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.User{}, id).Error; err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("error counting users: %w", err)
	}
	return count, nil
}

func (r *userRepository) Search(ctx context.Context, query string, limit, offset int) ([]domain.User, error) {
	var users []domain.User

	searchQuery := "%" + query + "%"

	dbQuery := r.db.WithContext(ctx).
		Where("username LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery)

	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}

	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}

	dbQuery = dbQuery.Order("id ASC")

	if err := dbQuery.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}

	return users, nil
}
