package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

func (r *AgentRepository) FindByID(id uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.First(&agent, id).Error
	return &agent, err
}

func (r *AgentRepository) List() ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.Order("id DESC").Find(&agents).Error
	return agents, err
}

func (r *AgentRepository) Update(agent *model.Agent) error {
	return r.db.Save(agent).Error
}

func (r *AgentRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("agent_id = ?", id).Delete(&model.AgentProject{}).Error; err != nil {
			return err
		}
		if err := tx.Where("agent_id = ?", id).Delete(&model.EnvironmentAgent{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Agent{}, id).Error
	})
}

func (r *AgentRepository) ListProjectIDs(agentID uint) ([]uint, error) {
	var links []model.AgentProject
	if err := r.db.Where("agent_id = ?", agentID).Order("project_id ASC").Find(&links).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.ProjectID)
	}
	return ids, nil
}

func (r *AgentRepository) SetProjectIDs(agentID uint, projectIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("agent_id = ?", agentID).Delete(&model.AgentProject{}).Error; err != nil {
			return err
		}
		if len(projectIDs) == 0 {
			return nil
		}
		links := make([]model.AgentProject, 0, len(projectIDs))
		for _, projectID := range projectIDs {
			links = append(links, model.AgentProject{
				AgentID:   agentID,
				ProjectID: projectID,
			})
		}
		return tx.Create(&links).Error
	})
}

// ListEnvironmentAgentIDs 按 sort_order 返回环境挂载的智能体 ID
func (r *AgentRepository) ListEnvironmentAgentIDs(environmentID uint) ([]uint, error) {
	var links []model.EnvironmentAgent
	if err := r.db.Where("environment_id = ?", environmentID).Order("sort_order ASC, agent_id ASC").Find(&links).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.AgentID)
	}
	return ids, nil
}

// SetEnvironmentAgentIDs 全量替换环境挂载；slice 顺序即 sort_order
func (r *AgentRepository) SetEnvironmentAgentIDs(environmentID uint, agentIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("environment_id = ?", environmentID).Delete(&model.EnvironmentAgent{}).Error; err != nil {
			return err
		}
		if len(agentIDs) == 0 {
			return nil
		}
		links := make([]model.EnvironmentAgent, 0, len(agentIDs))
		for i, agentID := range agentIDs {
			links = append(links, model.EnvironmentAgent{
				EnvironmentID: environmentID,
				AgentID:       agentID,
				SortOrder:     i,
			})
		}
		return tx.Create(&links).Error
	})
}

func (r *AgentRepository) DeleteEnvironmentLinks(environmentID uint) error {
	return r.db.Where("environment_id = ?", environmentID).Delete(&model.EnvironmentAgent{}).Error
}

func (r *AgentRepository) DeleteLinksByProjectID(projectID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&model.AgentProject{}).Error; err != nil {
			return err
		}
		return tx.Where(
			"environment_id IN (?)",
			tx.Model(&model.Environment{}).Select("id").Where("project_id = ?", projectID),
		).Delete(&model.EnvironmentAgent{}).Error
	})
}

// ListMountedAgentsForBuild 返回环境挂载且启用、且项目在范围内的智能体（按 sort_order）
func (r *AgentRepository) ListMountedAgentsForBuild(environmentID, projectID uint) ([]model.Agent, error) {
	var agents []model.Agent
	err := r.db.
		Joins("JOIN environment_agents ON environment_agents.agent_id = agents.id").
		Joins("JOIN agent_projects ON agent_projects.agent_id = agents.id AND agent_projects.project_id = ?", projectID).
		Where("environment_agents.environment_id = ? AND agents.enabled = ?", environmentID, true).
		Order("environment_agents.sort_order ASC, agents.id ASC").
		Find(&agents).Error
	return agents, err
}
