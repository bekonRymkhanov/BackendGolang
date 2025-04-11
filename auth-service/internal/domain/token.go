package domain

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// RefreshTokenClaims represents the claims in the refresh token
type RefreshTokenClaims struct {
	UserID uint `json:"user_id"`
}
