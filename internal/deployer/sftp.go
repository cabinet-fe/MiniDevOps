package deployer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPDeployer struct{}

func (d *SFTPDeployer) Deploy(ctx context.Context, opts DeployOptions) error {
	config, err := CreateSSHClientConfig(opts.Server)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", opts.Server.Host, opts.Server.Port)
	if opts.Server.Port == 0 {
		addr = opts.Server.Host + ":22"
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("sftp client: %w", err)
	}
	defer sftpClient.Close()

	remotePath := normalizeRemotePath(opts.Server, opts.RemotePath)
	if err := mkdirRecursive(sftpClient, remotePath); err != nil {
		return fmt.Errorf("create remote dir: %w", err)
	}

	return filepath.Walk(opts.SourceDir, func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(opts.SourceDir, localPath)
		if err != nil {
			return err
		}
		remoteFile := joinRemotePath(opts.Server, remotePath, relPath)

		if info.IsDir() {
			return mkdirRecursive(sftpClient, remoteFile)
		}

		return uploadFile(sftpClient, opts.Server, localPath, remoteFile, opts.Logger)
	})
}

func mkdirRecursive(c *sftp.Client, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	parts, cur := splitRemotePath(path)
	if cur != "" && !isWindowsDrive(cur) {
		if err := c.Mkdir(cur); err != nil {
			if _, statErr := c.Stat(cur); statErr != nil {
				return err
			}
		}
	}
	for _, p := range parts {
		if p == "" {
			continue
		}
		if cur == "" {
			cur = p
		} else if strings.HasSuffix(cur, ":") {
			cur = cur + `\` + p
		} else if strings.Contains(cur, `\`) || isWindowsDrive(cur) {
			cur = cur + `\` + p
		} else {
			cur = cur + "/" + p
		}
		if err := c.Mkdir(cur); err != nil {
			// Ignore if directory already exists
			if _, statErr := c.Stat(cur); statErr != nil {
				return err
			}
		}
	}
	return nil
}

func splitRemotePath(path string) ([]string, string) {
	root := ""
	trimmed := path
	if isWindowsDrive(trimmed) {
		root = trimmed[:2]
		trimmed = strings.TrimLeft(trimmed[2:], `\/`)
	}
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	return parts, root
}

func isWindowsDrive(path string) bool {
	return regexp.MustCompile(`^[A-Za-z]:`).MatchString(path)
}

func uploadFile(c *sftp.Client, server ServerInfo, localPath, remotePath string, logFn func(string)) error {
	if logFn != nil {
		logFn("Uploading: " + remotePath)
	}
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := mkdirRecursive(c, remoteDir(server, remotePath)); err != nil {
		return err
	}

	dst, err := c.Create(remotePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
