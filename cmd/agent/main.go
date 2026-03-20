package main

import (
	"archive/tar"
	"archive/zip"
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
	archiveFormat := normalizeArchiveFormat(r.Header.Get("X-Archive-Format"))

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := extractArchive(r.Body, targetPath, archiveFormat); err != nil {
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

func extractArchive(src io.Reader, targetDir, format string) error {
	tmpFile, err := os.CreateTemp("", "buildflow-upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, src); err != nil {
		return err
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return err
	}

	if normalizeArchiveFormat(format) == "zip" {
		return extractZip(tmpFile.Name(), targetDir)
	}
	return extractTarGz(tmpFile, targetDir)
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

func extractZip(srcPath, targetDir string) error {
	archive, err := zip.OpenReader(srcPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		targetPath := filepath.Join(targetDir, filepath.Clean(file.Name))
		relPath, err := filepath.Rel(targetDir, targetPath)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("illegal archive path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			reader.Close()
			return err
		}
		if _, err := io.Copy(dst, reader); err != nil {
			dst.Close()
			reader.Close()
			return err
		}
		if err := dst.Close(); err != nil {
			reader.Close()
			return err
		}
		if err := reader.Close(); err != nil {
			return err
		}
	}

	return nil
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

func normalizeArchiveFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "zip":
		return "zip"
	default:
		return "gzip"
	}
}
