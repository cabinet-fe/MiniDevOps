package service

import (
	"fmt"
	"time"

	"buildflow/internal/config"
	"buildflow/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims holds JWT claims.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles JWT generation and parsing.
type AuthService struct {
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	refreshKey  []byte // optional: separate key for refresh; defaults to same as secret
}

// NewAuthService creates an AuthService from config.
func NewAuthService(cfg *config.Config) (*AuthService, error) {
	if cfg == nil || cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	secret := []byte(cfg.JWT.Secret)

	accessTTL := 15 * time.Minute
	if cfg.JWT.AccessTTL != "" {
		d, err := time.ParseDuration(cfg.JWT.AccessTTL)
		if err != nil {
			return nil, fmt.Errorf("invalid access_ttl: %w", err)
		}
		accessTTL = d
	}

	refreshTTL := 7 * 24 * time.Hour
	if cfg.JWT.RefreshTTL != "" {
		d, err := time.ParseDuration(cfg.JWT.RefreshTTL)
		if err != nil {
			return nil, fmt.Errorf("invalid refresh_ttl: %w", err)
		}
		refreshTTL = d
	}

	return &AuthService{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		refreshKey: secret,
	}, nil
}

// GenerateTokenPair returns access and refresh tokens for the user.
func (s *AuthService) GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	now := time.Now()

	accessClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(s.secret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(s.refreshKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken parses and validates the access token, returns claims or error.
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ParseRefreshToken parses and validates a refresh token.
func (s *AuthService) ParseRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.refreshKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}
	return claims, nil
}
