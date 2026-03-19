package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", getenv("BUILDFLOW_AGENT_ADDR", ":9091"), "agent listen address")
	token := flag.String("token", getenv("BUILDFLOW_AGENT_TOKEN", ""), "agent bearer token")
	certFile := flag.String("tls-cert", getenv("BUILDFLOW_AGENT_TLS_CERT", ""), "TLS certificate path")
	keyFile := flag.String("tls-key", getenv("BUILDFLOW_AGENT_TLS_KEY", ""), "TLS private key path")
	flag.Parse()

	if strings.TrimSpace(*token) == "" {
		fmt.Fprintln(os.Stderr, "BUILDFLOW_AGENT_TOKEN or -token is required")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", withAuth(*token, healthzHandler))
	mux.HandleFunc("/upload", withAuth(*token, uploadHandler))
	mux.HandleFunc("/exec", withAuth(*token, execHandler))

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
	}

	if *certFile != "" && *keyFile != "" {
		if err := server.ListenAndServeTLS(*certFile, *keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "agent server failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "agent server failed: %v\n", err)
		os.Exit(1)
	}
}

func withAuth(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "ok (%s)", runtime.GOOS)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetPath := strings.TrimSpace(r.Header.Get("X-Target-Path"))
	if targetPath == "" {
		http.Error(w, "missing X-Target-Path header", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := extractTarGz(r.Body, targetPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "uploaded to %s", targetPath)
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Script  string `json:"script"`
		WorkDir string `json:"work_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Script) == "" {
		http.Error(w, "script is required", http.StatusBadRequest)
		return
	}

	cmd := commandForCurrentOS(req.Script)
	if strings.TrimSpace(req.WorkDir) != "" {
		cmd.Dir = req.WorkDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			http.Error(w, string(output), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(output)
}

func extractTarGz(src io.Reader, targetDir string) error {
	gzipReader, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, filepath.Clean(header.Name))
		relPath, err := filepath.Rel(targetDir, targetPath)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("illegal archive path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
	}
}

func commandForCurrentOS(script string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	}
	return exec.Command("sh", "-lc", script)
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
