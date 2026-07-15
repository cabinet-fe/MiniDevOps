package repository

import (
	"bedrock/internal/cicd/model"

	"gorm.io/gorm"
)

type ServerRepository struct{ db *gorm.DB }

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(s *model.Server) error {
	return r.db.Create(s).Error
}

func (r *ServerRepository) Update(s *model.Server) error {
	return r.db.Save(s).Error
}

func (r *ServerRepository) Delete(id uint) error {
	return r.db.Delete(&model.Server{}, id).Error
}

func (r *ServerRepository) FindByID(id uint) (*model.Server, error) {
	var s model.Server
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServerRepository) List(page, pageSize int, keyword, tag string) ([]model.Server, int64, error) {
	q := r.db.Model(&model.Server{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR host LIKE ?", like, like)
	}
	if tag != "" {
		q = q.Where("tags LIKE ?", "%"+tag+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Server
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *ServerRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Server{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ServerRepository) CountDeployTargets(serverID uint) (int64, error) {
	var n int64
	err := r.db.Model(&model.DeployTarget{}).Where("server_id = ?", serverID).Count(&n).Error
	return n, err
}
