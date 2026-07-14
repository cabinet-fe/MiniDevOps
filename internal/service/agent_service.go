package service

import (
	"errors"
	"strings"

	"bedrock/internal/model"
	"bedrock/internal/repository"
)

var validProxyKeys = map[string]bool{
	"opencode": true,
	"claude":   true,
	"reasonix": true,
}

type AgentService struct {
	repo        *repository.AgentRepository
	projectRepo *repository.ProjectRepository
}

func NewAgentService(repo *repository.AgentRepository, projectRepo *repository.ProjectRepository) *AgentService {
	return &AgentService{repo: repo, projectRepo: projectRepo}
}

func (s *AgentService) Create(agent *model.Agent, projectIDs []uint) error {
	if err := validateAgent(agent); err != nil {
		return err
	}
	ids := uniqueUintSlice(projectIDs)
	if err := s.validateProjectIDs(ids); err != nil {
		return err
	}
	if err := s.repo.Create(agent); err != nil {
		return err
	}
	if err := s.repo.SetProjectIDs(agent.ID, ids); err != nil {
		return err
	}
	agent.ProjectIDs = ids
	return nil
}

func (s *AgentService) GetByID(id uint) (*model.Agent, error) {
	agent, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	projectIDs, err := s.repo.ListProjectIDs(id)
	if err != nil {
		return nil, err
	}
	agent.ProjectIDs = projectIDs
	return agent, nil
}

func (s *AgentService) List() ([]model.Agent, error) {
	agents, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for i := range agents {
		projectIDs, err := s.repo.ListProjectIDs(agents[i].ID)
		if err != nil {
			return nil, err
		}
		agents[i].ProjectIDs = projectIDs
	}
	return agents, nil
}

func (s *AgentService) Update(agent *model.Agent, projectIDs []uint, syncProjects bool) error {
	if _, err := s.repo.FindByID(agent.ID); err != nil {
		return err
	}
	if err := validateAgent(agent); err != nil {
		return err
	}
	if err := s.repo.Update(agent); err != nil {
		return err
	}
	if syncProjects {
		ids := uniqueUintSlice(projectIDs)
		if err := s.validateProjectIDs(ids); err != nil {
			return err
		}
		if err := s.repo.SetProjectIDs(agent.ID, ids); err != nil {
			return err
		}
		agent.ProjectIDs = ids
	} else {
		ids, err := s.repo.ListProjectIDs(agent.ID)
		if err != nil {
			return err
		}
		agent.ProjectIDs = ids
	}
	return nil
}

func (s *AgentService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *AgentService) validateProjectIDs(ids []uint) error {
	for _, id := range ids {
		if _, err := s.projectRepo.FindByID(id); err != nil {
			return errors.New("项目不存在")
		}
	}
	return nil
}

func validateAgent(agent *model.Agent) error {
	agent.Name = strings.TrimSpace(agent.Name)
	if agent.Name == "" {
		return errors.New("名称不能为空")
	}
	if len(agent.Name) > 100 {
		return errors.New("名称过长")
	}
	agent.ProxyKey = strings.TrimSpace(agent.ProxyKey)
	if !validProxyKeys[agent.ProxyKey] {
		return errors.New("无效的代理类型，支持: opencode / claude / reasonix")
	}
	return nil
}

// IsValidProxyKey 供测试与 proxy 服务复用
func IsValidProxyKey(key string) bool {
	return validProxyKeys[key]
}
