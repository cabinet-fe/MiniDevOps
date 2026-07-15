package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	authhandler "bedrock/internal/auth/handler"
	authmiddleware "bedrock/internal/auth/middleware"
	authrepo "bedrock/internal/auth/repository"
	authservice "bedrock/internal/auth/service"
	"bedrock/internal/pkg"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	"bedrock/internal/platform/seed"
	rbacrepo "bedrock/internal/rbac/repository"
	rbacservice "bedrock/internal/rbac/service"
)

func setupAuthRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	const keyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := pkg.InitEncryption(keyHex); err != nil {
		t.Fatal(err)
	}

	gdb, err := db.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "auth.sqlite"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	})

	if err := migration.Up(context.Background(), gdb, "sqlite"); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureSuperAdmin(gdb, config.AdminConfig{
		Username:    "admin",
		Password:    "admin123",
		DisplayName: "管理员",
	}); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureRBACResources(gdb); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", AccessTTL: "1h", RefreshTTL: "24h"},
	}
	users := authrepo.NewUserRepository(gdb)
	roles := rbacrepo.NewRoleRepository(gdb)
	resources := rbacrepo.NewResourceRepository(gdb)
	permSvc := rbacservice.NewPermissionService(roles, resources)
	authSvc, err := authservice.NewAuthService(cfg, users, permSvc)
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	api := r.Group("/api/v1")
	h := authhandler.NewAuthHandler(authSvc)
	h.RegisterRoutes(api, authmiddleware.Auth(authSvc))
	return r
}

func TestLogin_passwordCipher(t *testing.T) {
	r := setupAuthRouter(t)

	cipher, err := pkg.EncryptLoginPasswordCipherForTest("admin123")
	if err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{
		"username":        "admin",
		"password_cipher": cipher,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
			User        struct {
				Username     string `json:"username"`
				IsSuperAdmin bool   `json:"is_super_admin"`
			} `json:"user"`
			Permissions []string `json:"permissions"`
			Menus       []any    `json:"menus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != 0 || resp.Data.AccessToken == "" {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
	if !resp.Data.User.IsSuperAdmin || resp.Data.User.Username != "admin" {
		t.Fatalf("user=%+v", resp.Data.User)
	}
	if len(resp.Data.Permissions) == 0 {
		t.Fatalf("super-admin should have permissions")
	}
	if len(resp.Data.Menus) == 0 {
		t.Fatalf("super-admin should have menus")
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req2.Header.Set("Authorization", "Bearer "+resp.Data.AccessToken)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", w2.Code, w2.Body.String())
	}
}
