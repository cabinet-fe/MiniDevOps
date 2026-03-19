package repository

import (
	"time"

	"buildflow/internal/model"

	"gorm.io/gorm"
)

type BuildRepository struct {
	db *gorm.DB
}

func NewBuildRepository(db *gorm.DB) *BuildRepository {
	return &BuildRepository{db: db}
}

func (r *BuildRepository) Create(build *model.Build) error {
	return r.db.Create(build).Error
}

func (r *BuildRepository) FindByID(id uint) (*model.Build, error) {
	var build model.Build
	err := r.db.First(&build, id).Error
	return &build, err
}

func (r *BuildRepository) List(projectID uint, environmentID *uint, page, pageSize int) ([]model.Build, int64, error) {
	var builds []model.Build
	var total int64
	query := r.db.Model(&model.Build{}).Where("project_id = ?", projectID)
	if environmentID != nil {
		query = query.Where("environment_id = ?", *environmentID)
	}
	query.Count(&total)
	err := query.Offset((page-1)*pageSize).Limit(pageSize).Order("created_at DESC").Find(&builds).Error
	return builds, total, err
}

func (r *BuildRepository) ListAll(page, pageSize int) ([]model.Build, int64, error) {
	var builds []model.Build
	var total int64
	r.db.Model(&model.Build{}).Count(&total)
	err := r.db.Offset((page-1)*pageSize).Limit(pageSize).Order("created_at DESC").Find(&builds).Error
	return builds, total, err
}

func (r *BuildRepository) GetNextBuildNumber(projectID uint) (int, error) {
	var maxNum *int
	err := r.db.Model(&model.Build{}).Where("project_id = ?", projectID).Select("MAX(build_number)").Scan(&maxNum).Error
	if err != nil {
		return 1, err
	}
	if maxNum == nil {
		return 1, nil
	}
	return *maxNum + 1, nil
}

func (r *BuildRepository) FindActiveBuilds() ([]model.Build, error) {
	var builds []model.Build
	err := r.db.Where("status IN ?", []string{"pending", "cloning", "building", "deploying"}).Find(&builds).Error
	return builds, err
}

func (r *BuildRepository) CountToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&model.Build{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

func (r *BuildRepository) CountByStatusInDays(days int) ([]struct {
	Date   string
	Status string
	Count  int64
}, error) {
	var results []struct {
		Date   string
		Status string
		Count  int64
	}
	from := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	err := r.db.Model(&model.Build{}).
		Select("DATE(created_at) as date, status, COUNT(*) as count").
		Where("created_at >= ?", from).
		Group("DATE(created_at), status").
		Find(&results).Error
	return results, err
}

func (r *BuildRepository) GetRecentBuilds(limit int) ([]model.Build, error) {
	var builds []model.Build
	err := r.db.Order("created_at DESC").Limit(limit).Find(&builds).Error
	return builds, err
}

func (r *BuildRepository) UpdateStatus(id uint, status string, fields map[string]interface{}) error {
	updates := map[string]interface{}{"status": status}
	for k, v := range fields {
		updates[k] = v
	}
	return r.db.Model(&model.Build{}).Where("id = ?", id).Updates(updates).Error
}

func (r *BuildRepository) FindArtifactsByProject(projectID uint) ([]model.Build, error) {
	var builds []model.Build
	err := r.db.Where("project_id = ? AND artifact_path != ''", projectID).
		Order("build_number DESC").
		Find(&builds).Error
	return builds, err
}

func (r *BuildRepository) DeleteByProjectID(projectID uint) error {
	return r.db.Where("project_id = ?", projectID).Delete(&model.Build{}).Error
}
