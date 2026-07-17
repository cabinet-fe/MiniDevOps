package repository

import (
	"bedrock/internal/dashboard/model"

	"gorm.io/gorm"
)

type DashboardRepository struct{ db *gorm.DB }

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) FindLayoutByUserID(userID uint) (*model.Layout, error) {
	var layout model.Layout
	if err := r.db.Where("user_id = ?", userID).First(&layout).Error; err != nil {
		return nil, err
	}
	return &layout, nil
}

func (r *DashboardRepository) CreateLayout(layout *model.Layout) error {
	return r.db.Create(layout).Error
}

func (r *DashboardRepository) UpdateLayout(layout *model.Layout) error {
	return r.db.Model(&model.Layout{}).Where("user_id = ?", layout.UserID).
		Updates(map[string]interface{}{"cards_json": layout.CardsJSON}).Error
}

func (r *DashboardRepository) CountRunsByStatus(status string) (int64, error) {
	var total int64
	err := r.db.Table("build_runs").Where("status = ?", status).Count(&total).Error
	return total, err
}

func (r *DashboardRepository) CountFinishedRuns() (total, success int64, err error) {
	err = r.db.Table("build_runs").
		Where("status IN ?", []string{"success", "failed", "cancelled", "interrupted"}).
		Count(&total).Error
	if err != nil {
		return 0, 0, err
	}
	err = r.db.Table("build_runs").Where("status = ?", "success").Count(&success).Error
	return total, success, err
}

func (r *DashboardRepository) ListRecentRuns(limit int) ([]model.RecentRun, error) {
	var rows []model.RecentRun
	err := r.db.Table("build_runs").
		Select("id, build_job_id, build_number, status, branch, created_at").
		Order("id DESC").Limit(limit).Scan(&rows).Error
	return rows, err
}

func (r *DashboardRepository) CountAgentRunsByStatus(status string) (int64, error) {
	var total int64
	err := r.db.Table("agent_runs").Where("status = ?", status).Count(&total).Error
	return total, err
}

func (r *DashboardRepository) CountAgentRunsByStatuses(statuses []string) (int64, error) {
	var total int64
	err := r.db.Table("agent_runs").Where("status IN ?", statuses).Count(&total).Error
	return total, err
}

func (r *DashboardRepository) CountFinishedAgentRuns() (total, success int64, err error) {
	err = r.db.Table("agent_runs").
		Where("status IN ?", []string{"success", "failed", "cancelled", "interrupted"}).
		Count(&total).Error
	if err != nil {
		return 0, 0, err
	}
	err = r.db.Table("agent_runs").Where("status = ?", "success").Count(&success).Error
	return total, success, err
}

func (r *DashboardRepository) ListRecentAgentRuns(limit int) ([]model.RecentAgentRun, error) {
	var rows []model.RecentAgentRun
	err := r.db.Table("agent_runs").
		Select("agent_runs.id, agent_runs.agent_id, ai_agents.name AS agent_name, agent_runs.trigger_type, agent_runs.status, agent_runs.created_at").
		Joins("LEFT JOIN ai_agents ON ai_agents.id = agent_runs.agent_id").
		Order("agent_runs.id DESC").Limit(limit).Scan(&rows).Error
	return rows, err
}
