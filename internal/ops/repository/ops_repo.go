package repository

import (
	"bedrock/internal/ops/model"

	"gorm.io/gorm"
)

type OpsRepository struct{ db *gorm.DB }

func NewOpsRepository(db *gorm.DB) *OpsRepository {
	return &OpsRepository{db: db}
}

func (r *OpsRepository) ListToolchains() ([]model.ToolchainDefinition, error) {
	var items []model.ToolchainDefinition
	err := r.db.Order("kind ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *OpsRepository) FindToolchain(id uint) (*model.ToolchainDefinition, error) {
	var item model.ToolchainDefinition
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpsRepository) CreateToolchain(item *model.ToolchainDefinition) error {
	return r.db.Create(item).Error
}

func (r *OpsRepository) UpdateToolchain(item *model.ToolchainDefinition) error {
	return r.db.Save(item).Error
}

func (r *OpsRepository) DeleteToolchain(id uint) error {
	return r.db.Delete(&model.ToolchainDefinition{}, id).Error
}

func (r *OpsRepository) ListSources() ([]model.InstallSource, error) {
	var items []model.InstallSource
	err := r.db.Order("priority ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *OpsRepository) FindSource(id uint) (*model.InstallSource, error) {
	var item model.InstallSource
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpsRepository) ListEnabledSources() ([]model.InstallSource, error) {
	var items []model.InstallSource
	err := r.db.Where("enabled = ?", true).Order("priority ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *OpsRepository) CreateSource(item *model.InstallSource) error {
	return r.db.Create(item).Error
}

func (r *OpsRepository) UpdateSource(item *model.InstallSource) error {
	return r.db.Save(item).Error
}

func (r *OpsRepository) DeleteSource(id uint) error {
	return r.db.Delete(&model.InstallSource{}, id).Error
}

func (r *OpsRepository) ListJobs(page, pageSize int, status string) ([]model.ToolchainInstallJob, int64, error) {
	q := r.db.Model(&model.ToolchainInstallJob{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.ToolchainInstallJob
	err := q.Preload("ToolchainDefinition").Preload("Source").Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *OpsRepository) FindJob(id uint) (*model.ToolchainInstallJob, error) {
	var item model.ToolchainInstallJob
	if err := r.db.Preload("ToolchainDefinition").Preload("Source").First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpsRepository) CreateJob(item *model.ToolchainInstallJob) error {
	return r.db.Create(item).Error
}

func (r *OpsRepository) UpdateJob(item *model.ToolchainInstallJob) error {
	return r.db.Save(item).Error
}

func (r *OpsRepository) ListJobsByStatuses(statuses ...string) ([]model.ToolchainInstallJob, error) {
	var items []model.ToolchainInstallJob
	if len(statuses) == 0 {
		return items, nil
	}
	err := r.db.Where("status IN ?", statuses).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *OpsRepository) MarkRunningInterrupted() (int64, error) {
	result := r.db.Model(&model.ToolchainInstallJob{}).Where("status = ?", model.JobRunning).
		Updates(map[string]interface{}{"status": model.JobInterrupted})
	return result.RowsAffected, result.Error
}
