package repository

import (
	"bedrock/internal/resource/model"

	"gorm.io/gorm"
)

type RepositoryRepository struct{ db *gorm.DB }

func NewRepositoryRepository(db *gorm.DB) *RepositoryRepository {
	return &RepositoryRepository{db: db}
}

func (r *RepositoryRepository) Create(repo *model.Repository) error {
	return r.db.Create(repo).Error
}

func (r *RepositoryRepository) Update(repo *model.Repository) error {
	return r.db.Save(repo).Error
}

func (r *RepositoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Repository{}, id).Error
}

func (r *RepositoryRepository) FindByID(id uint) (*model.Repository, error) {
	var repo model.Repository
	if err := r.db.First(&repo, id).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *RepositoryRepository) List(page, pageSize int, keyword string) ([]model.Repository, int64, error) {
	q := r.db.Model(&model.Repository{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR repo_url LIKE ? OR tags LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Repository
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *RepositoryRepository) CountJobs(repositoryID uint) (int64, error) {
	var n int64
	err := r.db.Table("build_jobs").Where("repository_id = ?", repositoryID).Count(&n).Error
	return n, err
}

func (r *RepositoryRepository) CountAgentBindings(repositoryID uint) (int64, error) {
	var n int64
	err := r.db.Table("ai_agent_repo_bindings").Where("repository_id = ?", repositoryID).Count(&n).Error
	return n, err
}
