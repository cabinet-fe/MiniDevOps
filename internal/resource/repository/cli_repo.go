package repository

import (
	"gorm.io/gorm"

	"bedrock/internal/resource/model"
)

type CLIRepository struct {
	db *gorm.DB
}

func NewCLIRepository(db *gorm.DB) *CLIRepository {
	return &CLIRepository{db: db}
}

func (r *CLIRepository) List() ([]model.CliRuntimeDefinition, error) {
	var items []model.CliRuntimeDefinition
	err := r.db.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *CLIRepository) FindByKey(key string) (*model.CliRuntimeDefinition, error) {
	var item model.CliRuntimeDefinition
	if err := r.db.Where("key = ?", key).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CLIRepository) Update(item *model.CliRuntimeDefinition) error {
	return r.db.Save(item).Error
}

func (r *CLIRepository) ListSources(cliKey string) ([]model.CliInstallSource, error) {
	var items []model.CliInstallSource
	q := r.db.Order("priority ASC, id ASC")
	if cliKey != "" {
		q = q.Where("cli_key = ?", cliKey)
	}
	err := q.Find(&items).Error
	return items, err
}

func (r *CLIRepository) ListEnabledSources(cliKey string) ([]model.CliInstallSource, error) {
	var items []model.CliInstallSource
	err := r.db.Where("cli_key = ? AND enabled = ?", cliKey, true).
		Order("priority ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *CLIRepository) FindSource(id uint) (*model.CliInstallSource, error) {
	var item model.CliInstallSource
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CLIRepository) CreateSource(item *model.CliInstallSource) error {
	return r.db.Create(item).Error
}

func (r *CLIRepository) UpdateSource(item *model.CliInstallSource) error {
	return r.db.Save(item).Error
}

func (r *CLIRepository) DeleteSource(id uint) error {
	return r.db.Delete(&model.CliInstallSource{}, id).Error
}
