package service

import (
	"errors"
	"fmt"
	"net"

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

	output, err := session.CombinedOutput("uname -a")
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %w", err)
	}
	return string(output), nil
}
