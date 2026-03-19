package deployer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	remotePath := opts.RemotePath
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
		remoteFile := filepath.Join(remotePath, filepath.ToSlash(relPath))

		if info.IsDir() {
			return mkdirRecursive(sftpClient, remoteFile)
		}

		return uploadFile(sftpClient, localPath, remoteFile, opts.Logger)
	})
}

func mkdirRecursive(c *sftp.Client, path string) error {
	path = filepath.ToSlash(path)
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	cur := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if cur == "" {
			cur = p
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

func uploadFile(c *sftp.Client, localPath, remotePath string, logFn func(string)) error {
	if logFn != nil {
		logFn("Uploading: " + remotePath)
	}
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := mkdirRecursive(c, filepath.Dir(remotePath)); err != nil {
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
