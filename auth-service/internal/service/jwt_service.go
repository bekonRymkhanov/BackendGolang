package service

import (
	"auth-service/internal/domain"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateToken(userID uint, username, role string) (string, error)
	GenerateRefreshToken(userID uint) (string, error)
	ValidateToken(tokenString string) (*domain.JWTClaims, error)
	ValidateRefreshToken(tokenString string) (*domain.RefreshTokenClaims, error)
}

type jwtService struct {
	secretKey        string
	refreshSecretKey string
	tokenExpiry      time.Duration
	refreshExpiry    time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(
	secretKey, refreshSecretKey string,
	tokenExpiry, refreshExpiry time.Duration,
) JWTService {
	return &jwtService{
		secretKey:        secretKey,
		refreshSecretKey: refreshSecretKey,
		tokenExpiry:      tokenExpiry,
		refreshExpiry:    refreshExpiry,
	}
}

// GenerateToken generates a new JWT token
func (s *jwtService) GenerateToken(userID uint, username, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(s.tokenExpiry).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// GenerateRefreshToken generates a new refresh token
func (s *jwtService) GenerateRefreshToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.refreshExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecretKey))
}

// ValidateToken validates a JWT token
func (s *jwtService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("invalid user_id in token")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return nil, errors.New("invalid username in token")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New("invalid role in token")
		}

		return &domain.JWTClaims{
			UserID:   uint(userID),
			Username: username,
			Role:     role,
		}, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates a refresh token
func (s *jwtService) ValidateRefreshToken(tokenString string) (*domain.RefreshTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.refreshSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("invalid user_id in token")
		}

		return &domain.RefreshTokenClaims{
			UserID: uint(userID),
		}, nil
	}

	return nil, errors.New("invalid refresh token")
}
