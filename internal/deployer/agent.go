package deployer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AgentDeployer struct{}

func (d *AgentDeployer) Deploy(ctx context.Context, opts DeployOptions) error {
	if strings.TrimSpace(opts.Server.AgentURL) == "" {
		return fmt.Errorf("agent url is required")
	}
	if strings.TrimSpace(opts.Server.AgentToken) == "" {
		return fmt.Errorf("agent token is required")
	}

	uploadURL, err := joinAgentURL(opts.Server.AgentURL, "upload")
	if err != nil {
		return err
	}

	archivePath, err := createArchive(opts.SourceDir)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, file)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+opts.Server.AgentToken)
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("X-Target-Path", normalizeRemotePath(opts.Server, opts.RemotePath))

	if info, statErr := file.Stat(); statErr == nil {
		req.ContentLength = info.Size()
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("agent upload failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("agent upload failed: %s", strings.TrimSpace(string(body)))
	}
	if opts.Logger != nil {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = "Agent upload completed"
		}
		opts.Logger(message)
	}
	return nil
}

func createArchive(sourceDir string) (string, error) {
	tmpFile, err := os.CreateTemp("", "buildflow-agent-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	gzipWriter := gzip.NewWriter(tmpFile)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	}); err != nil {
		tarWriter.Close()
		gzipWriter.Close()
		return "", err
	}

	if err := tarWriter.Close(); err != nil {
		gzipWriter.Close()
		return "", err
	}
	if err := gzipWriter.Close(); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}
