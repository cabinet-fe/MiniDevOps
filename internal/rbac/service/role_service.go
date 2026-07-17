package service

import (
	"errors"
	"fmt"
	"strings"

	"bedrock/internal/rbac"
	"bedrock/internal/rbac/model"
	"bedrock/internal/rbac/repository"
)

type RoleService struct {
	roles *repository.RoleRepository
}

func NewRoleService(roles *repository.RoleRepository) *RoleService {
	return &RoleService{roles: roles}
}

func (s *RoleService) Create(name, code, description string, permissions []string) (*model.Role, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" || code == "" {
		return nil, errors.New("名称与编码不能为空")
	}
	if err := validatePermissions(permissions); err != nil {
		return nil, err
	}
	role := &model.Role{Name: name, Code: code, Description: description}
	if err := s.roles.Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}
	if err := s.roles.ReplacePermissions(role.ID, permissions); err != nil {
		return nil, err
	}
	return s.roles.FindByID(role.ID)
}

func (s *RoleService) Get(id uint) (*model.Role, error) {
	return s.roles.FindByID(id)
}

func (s *RoleService) List(page, pageSize int) ([]model.Role, int64, error) {
	return s.roles.List(page, pageSize)
}

func (s *RoleService) Update(id uint, name, description string) (*model.Role, error) {
	role, err := s.roles.FindByID(id)
	if err != nil {
		return nil, err
	}
	if name = strings.TrimSpace(name); name != "" {
		role.Name = name
	}
	role.Description = description
	if err := s.roles.Update(role); err != nil {
		return nil, err
	}
	return s.roles.FindByID(id)
}

func (s *RoleService) Delete(id uint) error {
	return s.roles.Delete(id)
}

func (s *RoleService) SetPermissions(id uint, permissions []string) (*model.Role, error) {
	if _, err := s.roles.FindByID(id); err != nil {
		return nil, err
	}
	if err := validatePermissions(permissions); err != nil {
		return nil, err
	}
	if err := s.roles.ReplacePermissions(id, permissions); err != nil {
		return nil, err
	}
	return s.roles.FindByID(id)
}

func (s *RoleService) SetUserRoles(userID uint, roleIDs []uint) error {
	return s.roles.ReplaceUserRoles(userID, roleIDs)
}

func validatePermissions(permissions []string) error {
	for _, p := range permissions {
		if p == "" {
			continue
		}
		if _, _, ok := rbac.SplitPermission(p); !ok {
			return fmt.Errorf("无效权限码: %s", p)
		}
	}
	return nil
}
