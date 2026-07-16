package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"bedrock/internal/auth/middleware"
	"bedrock/internal/auth/service"
	"bedrock/internal/pkg"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// RegisterRoutes mounts auth endpoints under /api/v1.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	rg.POST("/auth/login", h.Login)
	rg.POST("/auth/refresh", h.Refresh)

	secured := rg.Group("", authMW)
	{
		secured.POST("/auth/logout", h.Logout)
		secured.GET("/auth/me", h.Me)
	}
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	maxAge := int(h.auth.RefreshTTL() / time.Second)
	if maxAge <= 0 {
		maxAge = int((7 * 24 * time.Hour) / time.Second)
	}
	c.SetSameSite(http.SameSiteLaxMode)
	// HttpOnly: browser JS must not read/write refresh_token.
	// Secure omitted so HTTP (intranet / IP) deployments still work.
	c.SetCookie(refreshCookieName, token, maxAge, "/api/v1/auth", "", false, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, "/api/v1/auth", "", false, true)
}

func (h *AuthHandler) readRefreshToken(c *gin.Context) string {
	if tok, err := c.Cookie(refreshCookieName); err == nil {
		if t := strings.TrimSpace(tok); t != "" {
			return t
		}
	}
	// Optional body fallback for non-browser clients (curl / smoke).
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return ""
	}
	return strings.TrimSpace(req.RefreshToken)
}

// POST /auth/login — prefers password_cipher; plaintext password allowed for debug only.
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
	} else if req.Password != "" {
		password = req.Password
	} else {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	user, err := h.auth.Authenticate(req.Username, password)
	if err != nil {
		pkg.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	accessToken, refreshToken, err := h.auth.GenerateTokenPair(user)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "生成Token失败")
		return
	}
	me, err := h.auth.Me(user.ID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "加载身份失败")
		return
	}
	h.setRefreshCookie(c, refreshToken)
	pkg.Success(c, gin.H{
		"access_token": accessToken,
		"user":         me.User,
		"permissions":  me.Permissions,
		"menus":        me.Menus,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearRefreshCookie(c)
	pkg.Success(c, nil)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken := h.readRefreshToken(c)
	if refreshToken == "" {
		h.clearRefreshCookie(c)
		pkg.Error(c, http.StatusUnauthorized, "请重新登录")
		return
	}
	claims, err := h.auth.ParseRefreshToken(refreshToken)
	if err != nil {
		h.clearRefreshCookie(c)
		pkg.Error(c, http.StatusUnauthorized, "Token已过期，请重新登录")
		return
	}
	user, err := h.auth.GetByID(claims.UserID)
	if err != nil || !user.IsActive {
		h.clearRefreshCookie(c)
		pkg.Error(c, http.StatusUnauthorized, "用户不存在或已被禁用")
		return
	}
	accessToken, newRefreshToken, err := h.auth.GenerateTokenPair(user)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "生成Token失败")
		return
	}
	h.setRefreshCookie(c, newRefreshToken)
	pkg.Success(c, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	payload, err := h.auth.Me(middleware.GetUserID(c))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	pkg.Success(c, payload)
}
