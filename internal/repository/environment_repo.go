package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

// EnvironmentListItem is one environment row for the global list API, with owning project name.
type EnvironmentListItem struct {
	model.Environment
	ProjectName string `json:"project_name"`
}

type EnvironmentRepository struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

func (r *EnvironmentRepository) Create(env *model.Environment) error {
	return r.db.Create(env).Error
}

func (r *EnvironmentRepository) FindByID(id uint) (*model.Environment, error) {
	var env model.Environment
	err := r.db.First(&env, id).Error
	return &env, err
}

func (r *EnvironmentRepository) ListByProjectID(projectID uint) ([]model.Environment, error) {
	var envs []model.Environment
	err := r.db.Where("project_id = ?", projectID).Order("sort_order ASC, id ASC").Find(&envs).Error
	return envs, err
}

func (r *EnvironmentRepository) DeleteByProjectID(projectID uint) error {
	return r.db.Where("project_id = ?", projectID).Delete(&model.Environment{}).Error
}

func (r *EnvironmentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Environment{}, id).Error
}

func (r *EnvironmentRepository) CountByDeployServerID(serverID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Environment{}).Where("deploy_server_id = ?", serverID).Count(&count).Error
	return count, err
}

func (r *EnvironmentRepository) Update(env *model.Environment) error {
	return r.db.Save(env).Error
}

func (r *EnvironmentRepository) ListCronEnabled() ([]model.Environment, error) {
	var envs []model.Environment
	err := r.db.Where("cron_enabled = ? AND cron_expression != ''", true).Find(&envs).Error
	return envs, err
}

// ListJoined returns environments joined with projects, with optional filters and pagination.
// When createdBy is non-nil, only environments whose project.created_by matches are returned (dev role).
func (r *EnvironmentRepository) ListJoined(page, pageSize int, projectID *uint, nameLike string, createdBy *uint) ([]EnvironmentListItem, int64, error) {
	base := r.db.Table("environments").Joins("INNER JOIN projects ON projects.id = environments.project_id")
	if createdBy != nil {
		base = base.Where("projects.created_by = ?", *createdBy)
	}
	if projectID != nil {
		base = base.Where("environments.project_id = ?", *projectID)
	}
	if nameLike != "" {
		base = base.Where("environments.name LIKE ?", "%"+nameLike+"%")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []EnvironmentListItem
	err := base.Select("environments.*, projects.name as project_name").
		Order("projects.name ASC, environments.sort_order ASC, environments.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
