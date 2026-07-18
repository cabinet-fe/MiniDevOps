package service

import (
	"errors"
	"fmt"
	"strings"

	"bedrock/internal/rbac"
	"bedrock/internal/rbac/model"
	"bedrock/internal/rbac/repository"
)

type MenuGroupService struct {
	groups *repository.MenuGroupRepository
}

func NewMenuGroupService(groups *repository.MenuGroupRepository) *MenuGroupService {
	return &MenuGroupService{groups: groups}
}

type CreateMenuGroupInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	RoutePrefix string `json:"route_prefix"`
	SortKey     int    `json:"sort_key"`
	Enabled     *bool  `json:"enabled"`
}

func (s *MenuGroupService) Create(in CreateMenuGroupInput) (*model.MenuGroup, error) {
	name := strings.TrimSpace(in.Name)
	code := strings.TrimSpace(in.Code)
	if name == "" || code == "" {
		return nil, errors.New("名称与编码不能为空")
	}
	if !rbac.ValidCode(code) {
		return nil, errors.New("code 不能包含 '.'")
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	g := &model.MenuGroup{
		Name: name, Code: code, RoutePrefix: strings.TrimSpace(in.RoutePrefix),
		SortKey: in.SortKey, Enabled: enabled,
	}
	if err := s.groups.Create(g); err != nil {
		return nil, fmt.Errorf("创建菜单分组失败: %w", err)
	}
	return s.groups.FindByID(g.ID)
}

func (s *MenuGroupService) Get(id uint) (*model.MenuGroup, error) {
	return s.groups.FindByID(id)
}

func (s *MenuGroupService) List() ([]model.MenuGroup, error) {
	return s.groups.List()
}

type UpdateMenuGroupInput struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	RoutePrefix *string `json:"route_prefix"`
	SortKey     *int    `json:"sort_key"`
	Enabled     *bool   `json:"enabled"`
}

func (s *MenuGroupService) Update(id uint, in UpdateMenuGroupInput) (*model.MenuGroup, error) {
	g, err := s.groups.FindByID(id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errors.New("名称不能为空")
		}
		g.Name = name
	}
	if in.Code != nil {
		code := strings.TrimSpace(*in.Code)
		if code == "" {
			return nil, errors.New("code 不能为空")
		}
		if !rbac.ValidCode(code) {
			return nil, errors.New("code 不能包含 '.'")
		}
		g.Code = code
	}
	if in.RoutePrefix != nil {
		g.RoutePrefix = strings.TrimSpace(*in.RoutePrefix)
	}
	if in.SortKey != nil {
		g.SortKey = *in.SortKey
	}
	if in.Enabled != nil {
		g.Enabled = *in.Enabled
	}
	if err := s.groups.Update(g); err != nil {
		return nil, err
	}
	return s.groups.FindByID(id)
}

func (s *MenuGroupService) Delete(id uint) error {
	n, err := s.groups.CountMenus(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("请先移除分组下的菜单")
	}
	return s.groups.Delete(id)
}
