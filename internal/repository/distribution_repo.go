package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type DistributionRepository struct {
	db *gorm.DB
}

func NewDistributionRepository(db *gorm.DB) *DistributionRepository {
	return &DistributionRepository{db: db}
}

func (r *DistributionRepository) Create(d *model.Distribution) error {
	return r.db.Create(d).Error
}

func (r *DistributionRepository) FindByID(id uint) (*model.Distribution, error) {
	var d model.Distribution
	err := r.db.First(&d, id).Error
	return &d, err
}

func (r *DistributionRepository) DeleteByEnvironmentID(envID uint) error {
	return r.db.Where("environment_id = ?", envID).Delete(&model.Distribution{}).Error
}

func (r *DistributionRepository) ListByEnvironmentID(envID uint) ([]model.Distribution, error) {
	var rows []model.Distribution
	err := r.db.Where("environment_id = ?", envID).Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *DistributionRepository) ReplaceForEnvironment(envID uint, items []model.Distribution) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("environment_id = ?", envID).Delete(&model.Distribution{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = 0
			items[i].EnvironmentID = envID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *DistributionRepository) CountByServerID(serverID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Distribution{}).Where("server_id = ?", serverID).Count(&count).Error
	return count, err
}
