//go:build !dev

package main

import (
	"bytes"
	"embed"
	"encoding/json"
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

func injectEncryptionKey(html []byte, keyHex string) []byte {
	keyJSON, err := json.Marshal(keyHex)
	if err != nil {
		keyJSON = []byte(`""`)
	}
	snippet := `<script>window.__BUILDFLOW_ENCRYPTION_KEY__=` + string(keyJSON) + `</script>`
	const marker = "</head>"
	idx := bytes.Index(html, []byte(marker))
	if idx < 0 {
		return html
	}
	out := make([]byte, 0, len(html)+len(snippet))
	out = append(out, html[:idx]...)
	out = append(out, snippet...)
	out = append(out, html[idx:]...)
	return out
}

func serveSPA(r *gin.Engine, encryptionKeyHex string) {
	distFS, err := fs.Sub(webFS, "dist")
	if err != nil {
		return
	}
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		return
	}
	staticServer := http.FileServer(http.FS(distFS))

	serveIndex := func(c *gin.Context) {
		html := injectEncryptionKey(indexHTML, encryptionKeyHex)
		c.Data(http.StatusOK, "text/html; charset=utf-8", html)
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}

		trimmedPath := strings.TrimPrefix(path, "/")
		if trimmedPath == "" || trimmedPath == "index.html" {
			serveIndex(c)
			return
		}

		if fileInfo, err := fs.Stat(distFS, trimmedPath); err == nil && !fileInfo.IsDir() {
			c.Request.URL.Path = path
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		serveIndex(c)
	})
}
