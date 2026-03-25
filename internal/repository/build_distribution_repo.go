package repository

import (
	"errors"

	"buildflow/internal/model"

	"gorm.io/gorm"
)

type BuildDistributionRepository struct {
	db *gorm.DB
}

func NewBuildDistributionRepository(db *gorm.DB) *BuildDistributionRepository {
	return &BuildDistributionRepository{db: db}
}

func (r *BuildDistributionRepository) Create(row *model.BuildDistribution) error {
	return r.db.Create(row).Error
}

func (r *BuildDistributionRepository) Upsert(row *model.BuildDistribution) error {
	var existing model.BuildDistribution
	err := r.db.Where("build_id = ? AND distribution_id = ?", row.BuildID, row.DistributionID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(row).Error
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	return r.db.Model(&model.BuildDistribution{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
		"status":         row.Status,
		"error_message":  row.ErrorMessage,
		"started_at":     row.StartedAt,
		"finished_at":    row.FinishedAt,
	}).Error
}

func (r *BuildDistributionRepository) ListByBuildID(buildID uint) ([]model.BuildDistribution, error) {
	var rows []model.BuildDistribution
	err := r.db.Where("build_id = ?", buildID).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *BuildDistributionRepository) DeleteByBuildID(buildID uint) error {
	return r.db.Where("build_id = ?", buildID).Delete(&model.BuildDistribution{}).Error
}
