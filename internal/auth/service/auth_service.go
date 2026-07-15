package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"bedrock/internal/auth/model"
	"bedrock/internal/auth/repository"
	"bedrock/internal/pkg"
	"bedrock/internal/platform/config"
	rbacmodel "bedrock/internal/rbac/model"
	rbacservice "bedrock/internal/rbac/service"
)

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims holds JWT claims.
type Claims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

// AuthService handles JWT generation/parsing and login orchestration.
type AuthService struct {
	users      *repository.UserRepository
	perm       *rbacservice.PermissionService
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	refreshKey []byte
}

func NewAuthService(cfg *config.Config, users *repository.UserRepository, perm *rbacservice.PermissionService) (*AuthService, error) {
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
		users:      users,
		perm:       perm,
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		refreshKey: secret,
	}, nil
}

func (s *AuthService) GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	now := time.Now()

	accessClaims := Claims{
		UserID:       user.ID,
		Username:     user.Username,
		IsSuperAdmin: user.IsSuperAdmin,
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
		UserID:       user.ID,
		Username:     user.Username,
		IsSuperAdmin: user.IsSuperAdmin,
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

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	return s.parse(tokenString, s.secret)
}

func (s *AuthService) ParseRefreshToken(tokenString string) (*Claims, error) {
	return s.parse(tokenString, s.refreshKey)
}

func (s *AuthService) parse(tokenString string, key []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
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

func (s *AuthService) Authenticate(username, password string) (*model.User, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if !user.IsActive {
		return nil, errors.New("账户已被禁用")
	}
	if !pkg.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}
	return user, nil
}

func (s *AuthService) GetByID(id uint) (*model.User, error) {
	return s.users.FindByID(id)
}

// MePayload is returned by GET /auth/me.
type MePayload struct {
	User        *model.User          `json:"user"`
	Permissions []string             `json:"permissions"`
	Menus       []rbacmodel.MenuNode `json:"menus"`
}

func (s *AuthService) Me(userID uint) (*MePayload, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New("账户已被禁用")
	}
	perms := []string{}
	menus := []rbacmodel.MenuNode{}
	if s.perm != nil {
		perms, err = s.perm.ResolvePermissions(userID, user.IsSuperAdmin)
		if err != nil {
			return nil, err
		}
		menus, err = s.perm.TrimMenus(userID, user.IsSuperAdmin)
		if err != nil {
			return nil, err
		}
	}
	if perms == nil {
		perms = []string{}
	}
	if menus == nil {
		menus = []rbacmodel.MenuNode{}
	}
	return &MePayload{
		User:        user,
		Permissions: perms,
		Menus:       menus,
	}, nil
}
