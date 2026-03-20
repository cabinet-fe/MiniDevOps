package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *model.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) FindByID(id uint) (*model.Project, error) {
	var project model.Project
	err := r.db.Preload("Environments").First(&project, id).Error
	return &project, err
}

func (r *ProjectRepository) FindByName(name string) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("name = ?", name).First(&project).Error
	return &project, err
}

func (r *ProjectRepository) List(page, pageSize int, createdBy *uint) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64
	query := r.db.Model(&model.Project{})
	if createdBy != nil {
		query = query.Where("created_by = ?", *createdBy)
	}
	query.Count(&total)
	err := query.Preload("Environments").Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&projects).Error
	return projects, total, err
}

func (r *ProjectRepository) ListAll(createdBy *uint) ([]model.Project, error) {
	var projects []model.Project
	query := r.db.Preload("Environments").Order("name ASC")
	if createdBy != nil {
		query = query.Where("created_by = ?", *createdBy)
	}
	err := query.Find(&projects).Error
	return projects, err
}

// ListProjectTagsWithEnvCounts returns each project's tags string and environment count for dashboard tag summary (avoids loading full Environment rows).
func (r *ProjectRepository) ListProjectTagsWithEnvCounts(createdBy *uint) ([]struct {
	Tags     string
	EnvCount int64 `gorm:"column:env_count"`
}, error) {
	var rows []struct {
		Tags     string
		EnvCount int64 `gorm:"column:env_count"`
	}
	q := r.db.Model(&model.Project{}).
		Select("projects.tags, COUNT(environments.id) as env_count").
		Joins("LEFT JOIN environments ON environments.project_id = projects.id").
		Group("projects.id").
		Order("projects.name ASC")
	if createdBy != nil {
		q = q.Where("projects.created_by = ?", *createdBy)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

func (r *ProjectRepository) Update(project *model.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Delete(&model.Project{}, id).Error
}

func (r *ProjectRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Project{}).Count(&count).Error
	return count, err
}
