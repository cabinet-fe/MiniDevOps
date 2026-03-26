package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// WebSocketCheckOrigin returns whether the request Origin is allowed for WebSocket upgrades.
// Empty AllowOrigins matches CORSGin: allow any origin (typical dev).
func WebSocketCheckOrigin(cfg CORSConfig, r *http.Request) bool {
	allowAll := len(cfg.AllowOrigins) == 0
	if allowAll {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	for _, o := range cfg.AllowOrigins {
		if o == origin || o == "*" {
			return true
		}
	}
	return false
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowOrigins []string // Empty means allow all in dev
	AllowHeaders []string
	AllowMethods []string
}

// DefaultCORSConfig returns a config with common defaults.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: nil, // Allow all when empty (dev mode)
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
	}
}

// CORS returns a CORS middleware. In dev (AllowOrigins empty), allows all origins.
// In prod, only allows configured origins.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	allowAll := len(cfg.AllowOrigins) == 0
	origins := make(map[string]bool)
	for _, o := range cfg.AllowOrigins {
		origins[o] = true
	}

	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	if allowHeaders == "" {
		allowHeaders = "Origin, Content-Type, Accept, Authorization"
	}
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	if allowMethods == "" {
		allowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}

			if allowAll || origins[origin] || origins["*"] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSGin returns a Gin middleware for CORS.
func CORSGin(cfg CORSConfig) func(*gin.Context) {
	allowAll := len(cfg.AllowOrigins) == 0
	origins := make(map[string]bool)
	for _, o := range cfg.AllowOrigins {
		origins[o] = true
	}

	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	if allowHeaders == "" {
		allowHeaders = "Origin, Content-Type, Accept, Authorization"
	}
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	if allowMethods == "" {
		allowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		if allowAll || origins[origin] || origins["*"] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Headers", allowHeaders)
		c.Header("Access-Control-Allow-Methods", allowMethods)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
