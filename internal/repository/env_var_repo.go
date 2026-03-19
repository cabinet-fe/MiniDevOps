package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type EnvVarRepository struct {
	db *gorm.DB
}

func NewEnvVarRepository(db *gorm.DB) *EnvVarRepository {
	return &EnvVarRepository{db: db}
}

func (r *EnvVarRepository) ListByEnvironmentID(environmentID uint) ([]model.EnvVar, error) {
	var vars []model.EnvVar
	err := r.db.Where("environment_id = ?", environmentID).Order("id ASC").Find(&vars).Error
	return vars, err
}

func (r *EnvVarRepository) FindByID(id uint) (*model.EnvVar, error) {
	var envVar model.EnvVar
	err := r.db.First(&envVar, id).Error
	return &envVar, err
}

func (r *EnvVarRepository) Create(envVar *model.EnvVar) error {
	return r.db.Create(envVar).Error
}

func (r *EnvVarRepository) Update(envVar *model.EnvVar) error {
	return r.db.Save(envVar).Error
}

func (r *EnvVarRepository) Delete(id uint) error {
	return r.db.Delete(&model.EnvVar{}, id).Error
}

func (r *EnvVarRepository) DeleteByEnvironmentID(environmentID uint) error {
	return r.db.Where("environment_id = ?", environmentID).Delete(&model.EnvVar{}).Error
}

func (r *EnvVarRepository) DeleteByProjectID(projectID uint) error {
	return r.db.Where(
		"environment_id IN (?)",
		r.db.Model(&model.Environment{}).Select("id").Where("project_id = ?", projectID),
	).Delete(&model.EnvVar{}).Error
}
