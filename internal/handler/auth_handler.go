package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"buildflow/internal/middleware"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type AuthHandler struct {
	userService *service.UserService
	authService *service.AuthService
}

func NewAuthHandler(us *service.UserService, as *service.AuthService) *AuthHandler {
	return &AuthHandler{userService: us, authService: as}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username       string `json:"username" binding:"required"`
		Password       string `json:"password"`
		PasswordCipher string `json:"password_cipher"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	var password string
	if strings.TrimSpace(req.PasswordCipher) != "" {
		p, err := pkg.DecryptLoginPasswordCipher(strings.TrimSpace(req.PasswordCipher))
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, "登录参数无效")
			return
		}
		password = p
	} else {
		if req.Password == "" {
			pkg.Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		password = req.Password
	}
	user, err := h.userService.Authenticate(req.Username, password)
	if err != nil {
		pkg.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	accessToken, refreshToken, err := h.authService.GenerateTokenPair(user)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "生成Token失败")
		return
	}
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", false, true)
	pkg.Success(c, gin.H{
		"access_token": accessToken,
		"user":         user,
	})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	pkg.Success(c, nil)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		pkg.Error(c, http.StatusUnauthorized, "请重新登录")
		return
	}
	claims, err := h.authService.ParseRefreshToken(refreshToken)
	if err != nil {
		pkg.Error(c, http.StatusUnauthorized, "Token已过期，请重新登录")
		return
	}
	userID := claims.UserID
	user, err := h.userService.GetByID(userID)
	if err != nil || !user.IsActive {
		pkg.Error(c, http.StatusUnauthorized, "用户不存在或已被禁用")
		return
	}
	accessToken, newRefreshToken, err := h.authService.GenerateTokenPair(user)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "生成Token失败")
		return
	}
	c.SetCookie("refresh_token", newRefreshToken, 7*24*3600, "/", "", false, true)
	pkg.Success(c, gin.H{"access_token": accessToken, "user": user})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.userService.GetByID(userID)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	pkg.Success(c, user)
}

// PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Avatar      string `json:"avatar"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.OldPassword != "" && req.NewPassword != "" {
		if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
			pkg.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	if err := h.userService.UpdateProfile(userID, req.DisplayName, req.Email, req.Avatar); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	pkg.Success(c, nil)
}
