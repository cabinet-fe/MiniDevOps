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

// applyBuildProjectScope restricts builds to project_id IN filter when filter is non-nil.
// nil filter = no restriction; empty slice = match nothing.
func applyBuildProjectScope(db *gorm.DB, projectIDFilter []uint) *gorm.DB {
	if projectIDFilter == nil {
		return db
	}
	if len(projectIDFilter) == 0 {
		return db.Where("1 = 0")
	}
	return db.Where("project_id IN ?", projectIDFilter)
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
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&builds).Error
	return builds, total, err
}

// ListAll lists builds; projectIDFilter nil = all projects, empty slice = none.
func (r *BuildRepository) ListAll(page, pageSize int, projectIDFilter []uint) ([]model.Build, int64, error) {
	var builds []model.Build
	var total int64
	q := applyBuildProjectScope(r.db.Model(&model.Build{}), projectIDFilter)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := applyBuildProjectScope(r.db.Model(&model.Build{}), projectIDFilter).
		Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&builds).Error
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

func (r *BuildRepository) FindActiveBuilds(projectIDFilter []uint) ([]model.Build, error) {
	var builds []model.Build
	q := r.db.Where(
		"status IN ? OR (status = ? AND distribution_summary IN ?)",
		[]string{"pending", "cloning", "building", "deploying"},
		"success",
		[]string{"pending", "running"},
	)
	q = applyBuildProjectScope(q, projectIDFilter)
	err := q.Find(&builds).Error
	return builds, err
}

// FindByEnvironmentID returns all builds for an environment (any status).
func (r *BuildRepository) FindByEnvironmentID(environmentID uint) ([]model.Build, error) {
	var builds []model.Build
	err := r.db.Where("environment_id = ?", environmentID).Find(&builds).Error
	return builds, err
}

func (r *BuildRepository) DeleteByEnvironmentID(environmentID uint) error {
	return r.db.Where("environment_id = ?", environmentID).Delete(&model.Build{}).Error
}

// MarkInterruptedBuilds sets non-terminal builds to failed (e.g. after process crash).
func (r *BuildRepository) MarkInterruptedBuilds(errMsg string) (int64, error) {
	builds, err := r.FindActiveBuilds(nil)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	var n int64
	for _, b := range builds {
		if b.Status == "success" && (b.DistributionSummary == "running" || b.DistributionSummary == "pending") {
			if err := r.UpdateStatus(b.ID, "success", map[string]interface{}{
				"current_stage":          "success",
				"distribution_summary":   "cancelled",
				"redistribute_filter_json": "",
			}); err != nil {
				return n, err
			}
			n++
			continue
		}
		stage := b.CurrentStage
		if stage == "" {
			stage = b.Status
		}
		durationMs := b.DurationMs
		if b.StartedAt != nil {
			durationMs = now.Sub(*b.StartedAt).Milliseconds()
		}
		fields := map[string]interface{}{
			"current_stage":  stage,
			"error_message":  errMsg,
			"finished_at":    &now,
			"duration_ms":    durationMs,
		}
		if err := r.UpdateStatus(b.ID, "failed", fields); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (r *BuildRepository) CountActive(projectIDFilter []uint) (int64, error) {
	var count int64
	q := r.db.Model(&model.Build{}).
		Where(
			"status IN ? OR (status = ? AND distribution_summary IN ?)",
			[]string{"pending", "cloning", "building", "deploying"},
			"success",
			[]string{"pending", "running"},
		)
	q = applyBuildProjectScope(q, projectIDFilter)
	err := q.Count(&count).Error
	return count, err
}

// CountSuccessRateInDays returns success count and total builds in the last `days` days (from midnight).
func (r *BuildRepository) CountSuccessRateInDays(days int, projectIDFilter []uint) (success int64, total int64, err error) {
	from := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	var result struct {
		Total   int64 `gorm:"column:total"`
		Success int64 `gorm:"column:success"`
	}
	q := applyBuildProjectScope(r.db.Model(&model.Build{}), projectIDFilter).Where("created_at >= ?", from)
	err = q.Select("COUNT(*) as total, COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) as success").
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Success, result.Total, nil
}

func (r *BuildRepository) CountToday(projectIDFilter []uint) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	q := applyBuildProjectScope(r.db.Model(&model.Build{}), projectIDFilter).Where("created_at >= ?", today)
	err := q.Count(&count).Error
	return count, err
}

func (r *BuildRepository) CountByStatusInDays(days int, projectIDFilter []uint) ([]struct {
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
	q := applyBuildProjectScope(r.db.Model(&model.Build{}), projectIDFilter).
		Select("DATE(created_at) as date, status, COUNT(*) as count").
		Where("created_at >= ?", from).
		Group("DATE(created_at), status")
	err := q.Find(&results).Error
	return results, err
}

func (r *BuildRepository) GetRecentBuilds(limit int, projectIDFilter []uint) ([]model.Build, error) {
	var builds []model.Build
	q := applyBuildProjectScope(r.db.Model(&model.Build{}), projectIDFilter).Order("created_at DESC").Limit(limit)
	err := q.Find(&builds).Error
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
