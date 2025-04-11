// internal/handler/auth_handler.go
package handler

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var input domain.UserRegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUsernameExists) || errors.Is(err, service.ErrEmailExists) {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "validation failed") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var input domain.UserLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, service.ErrAccountDeactivated) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
		return
	}

	refreshToken := parts[1]
	token, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, service.ErrAccountDeactivated) {
			status = http.StatusForbidden
		} else if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

// Me returns the authenticated user
func (h *AuthHandler) Me(c *gin.Context) {
	// Из context берем user-а
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userObj.(*domain.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateUser handles user profile updates
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	var input domain.UserUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Из context берем user-а
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userObj.(*domain.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	updatedUser, err := h.authService.UpdateUser(c.Request.Context(), user.ID, input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrEmailExists) {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
		} else if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser.ToResponse())
}

// Logout is a placeholder for client-side logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: надо будет дописать что бы добавлять токен в блэклист или типо того для большой безопастности, а так можно оставить на стороне клиента только

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// ListUsers returns a list of all users (admin only)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	// Парсим конфиги пагинаций(можно будет изменять на стороне клиента)
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	users, total, err := h.authService.GetUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	userResponses := make([]domain.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := domain.UsersListResponse{
		Users:      userResponses,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetUser returns a specific user (admin only)
func (h *AuthHandler) GetUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteUser deletes a user (admin only)
func (h *AuthHandler) DeleteUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.authService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully deleted"})
}

// ActivateUser activates a user account (admin only)
func (h *AuthHandler) ActivateUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.authService.ChangeUserStatus(c.Request.Context(), uint(id), true); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully activated"})
}

// DeactivateUser deactivates a user account (admin only)
func (h *AuthHandler) DeactivateUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.authService.ChangeUserStatus(c.Request.Context(), uint(id), false); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully deactivated"})
}

// ChangeUserRole changes a user's role (admin only)
func (h *AuthHandler) ChangeUserRole(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var input struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ChangeUserRole(c.Request.Context(), uint(id), input.Role); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid role") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role successfully updated"})
}
