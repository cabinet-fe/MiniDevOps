package service

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"

	"golang.org/x/crypto/ssh"
)

type ServerService struct {
	repo    *repository.ServerRepository
	envRepo *repository.EnvironmentRepository
}

func NewServerService(repo *repository.ServerRepository, envRepo *repository.EnvironmentRepository) *ServerService {
	return &ServerService{repo: repo, envRepo: envRepo}
}

func (s *ServerService) Create(server *model.Server) error {
	normalizeServer(server)
	if err := validateServer(server); err != nil {
		return err
	}
	if server.Password != "" {
		enc, err := pkg.Encrypt(server.Password)
		if err != nil {
			return err
		}
		server.Password = enc
	}
	if server.PrivateKey != "" {
		enc, err := pkg.Encrypt(server.PrivateKey)
		if err != nil {
			return err
		}
		server.PrivateKey = enc
	}
	if server.AgentToken != "" {
		enc, err := pkg.Encrypt(server.AgentToken)
		if err != nil {
			return err
		}
		server.AgentToken = enc
	}
	return s.repo.Create(server)
}

func (s *ServerService) GetByID(id uint) (*model.Server, error) {
	server, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if server.Password != "" {
		dec, err := pkg.Decrypt(server.Password)
		if err != nil {
			return nil, err
		}
		server.Password = dec
	}
	if server.PrivateKey != "" {
		dec, err := pkg.Decrypt(server.PrivateKey)
		if err != nil {
			return nil, err
		}
		server.PrivateKey = dec
	}
	if server.AgentToken != "" {
		dec, err := pkg.Decrypt(server.AgentToken)
		if err != nil {
			return nil, err
		}
		server.AgentToken = dec
	}
	normalizeServer(server)
	return server, nil
}

func (s *ServerService) List(page, pageSize int, tag, role string) ([]model.Server, int64, error) {
	servers, total, err := s.repo.List(page, pageSize, tag)
	if err != nil {
		return nil, 0, err
	}
	if role == "dev" {
		// Return simplified list without credentials
		for i := range servers {
			servers[i].Password = ""
			servers[i].PrivateKey = ""
			servers[i].AgentToken = ""
		}
	}
	return servers, total, nil
}

func (s *ServerService) Update(server *model.Server) error {
	existing, err := s.repo.FindByID(server.ID)
	if err != nil {
		return err
	}
	// Re-encrypt if password/private_key changed (compare with existing stored values)
	// We cannot compare decrypted - assume if incoming non-empty is different from stored
	// For simplicity: if server.Password is non-empty and not hex (encrypted), encrypt it
	if server.Password != "" && server.Password != existing.Password {
		enc, err := pkg.Encrypt(server.Password)
		if err != nil {
			return err
		}
		server.Password = enc
	} else if server.Password == "" {
		server.Password = existing.Password
	}
	if server.PrivateKey != "" && server.PrivateKey != existing.PrivateKey {
		enc, err := pkg.Encrypt(server.PrivateKey)
		if err != nil {
			return err
		}
		server.PrivateKey = enc
	} else if server.PrivateKey == "" {
		server.PrivateKey = existing.PrivateKey
	}
	if server.AgentToken != "" && server.AgentToken != existing.AgentToken {
		enc, err := pkg.Encrypt(server.AgentToken)
		if err != nil {
			return err
		}
		server.AgentToken = enc
	} else if server.AgentToken == "" {
		server.AgentToken = existing.AgentToken
	}
	normalizeServer(server)
	if err := validateServer(server); err != nil {
		return err
	}
	return s.repo.Update(server)
}

func (s *ServerService) Delete(id uint) error {
	count, err := s.envRepo.CountByDeployServerID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该服务器正在被环境引用，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *ServerService) TestConnection(id uint) (string, error) {
	server, err := s.GetByID(id)
	if err != nil {
		return "", err
	}
	if server.AuthType == "agent" {
		return testAgentConnection(server)
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	if server.Port == 0 {
		addr = server.Host + ":22"
	}

	var authMethods []ssh.AuthMethod
	if server.AuthType == "key" && server.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(server.PrivateKey))
		if err != nil {
			return "", fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if server.AuthType == "password" && server.Password != "" {
		authMethods = append(authMethods, ssh.Password(server.Password))
	}
	if len(authMethods) == 0 {
		return "", errors.New("无法认证：未配置密码或私钥")
	}

	config := &ssh.ClientConfig{
		User: server.Username,
		Auth: authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	testCommand := "uname -a"
	if server.OSType == "windows" {
		testCommand = "cmd /c ver"
	}

	output, err := session.CombinedOutput(testCommand)
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %w", err)
	}
	return string(output), nil
}

func normalizeServer(server *model.Server) {
	server.OSType = strings.ToLower(strings.TrimSpace(server.OSType))
	if server.OSType == "" {
		server.OSType = "linux"
	}
	if server.OSType != "windows" {
		server.OSType = "linux"
	}

	server.AuthType = strings.ToLower(strings.TrimSpace(server.AuthType))
	server.AgentURL = strings.TrimSpace(server.AgentURL)
	server.Username = strings.TrimSpace(server.Username)
	server.Host = strings.TrimSpace(server.Host)

	if server.AuthType == "agent" {
		if parsed, err := url.Parse(server.AgentURL); err == nil && parsed.Hostname() != "" {
			if server.Host == "" {
				server.Host = parsed.Hostname()
			}
			if server.Port == 0 {
				if port := parsed.Port(); port != "" {
					if parsedPort, convErr := strconv.Atoi(port); convErr == nil {
						server.Port = parsedPort
					}
				} else if parsed.Scheme == "https" {
					server.Port = 443
				} else if parsed.Scheme == "http" {
					server.Port = 80
				}
			}
		}
		return
	}

	if server.Port == 0 {
		server.Port = 22
	}
}

func validateServer(server *model.Server) error {
	switch server.AuthType {
	case "password":
		if server.Host == "" {
			return errors.New("主机地址不能为空")
		}
		if server.Username == "" {
			return errors.New("用户名不能为空")
		}
		if strings.TrimSpace(server.Password) == "" {
			return errors.New("密码不能为空")
		}
	case "key":
		if server.Host == "" {
			return errors.New("主机地址不能为空")
		}
		if server.Username == "" {
			return errors.New("用户名不能为空")
		}
		if strings.TrimSpace(server.PrivateKey) == "" {
			return errors.New("SSH 私钥不能为空")
		}
	case "agent":
		if server.AgentURL == "" {
			return errors.New("Agent URL 不能为空")
		}
		if _, err := url.ParseRequestURI(server.AgentURL); err != nil {
			return fmt.Errorf("Agent URL 不合法: %w", err)
		}
		if strings.TrimSpace(server.AgentToken) == "" {
			return errors.New("Agent Token 不能为空")
		}
	default:
		return errors.New("不支持的认证方式")
	}
	return nil
}

func testAgentConnection(server *model.Server) (string, error) {
	healthURL, err := joinAgentURL(server.AgentURL, "healthz")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, healthURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+server.AgentToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Agent 连接失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return "", fmt.Errorf("Agent 健康检查失败: %s", message)
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		message = "agent healthy"
	}
	return message, nil
}

func joinAgentURL(baseURL, path string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("解析 Agent URL 失败: %w", err)
	}
	joined, err := url.JoinPath(parsed.String(), path)
	if err != nil {
		return "", fmt.Errorf("拼接 Agent URL 失败: %w", err)
	}
	return joined, nil
}
