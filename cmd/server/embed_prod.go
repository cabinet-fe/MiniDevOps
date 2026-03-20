//go:build !dev

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var webFS embed.FS

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func serveSPA(r *gin.Engine) {
	distFS, err := fs.Sub(webFS, "dist")
	if err != nil {
		return
	}
	staticServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}

		trimmedPath := strings.TrimPrefix(path, "/")
		if trimmedPath == "" {
			c.Request.URL.Path = "/"
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		if fileInfo, err := fs.Stat(distFS, trimmedPath); err == nil && !fileInfo.IsDir() {
			c.Request.URL.Path = path
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Request.URL.Path = "/"
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}
