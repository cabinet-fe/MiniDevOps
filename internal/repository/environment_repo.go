package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

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
