package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(server *model.Server) error {
	return r.db.Create(server).Error
}

func (r *ServerRepository) FindByID(id uint) (*model.Server, error) {
	var server model.Server
	err := r.db.First(&server, id).Error
	return &server, err
}

func (r *ServerRepository) List(page, pageSize int, tag string) ([]model.Server, int64, error) {
	var servers []model.Server
	var total int64
	query := r.db.Model(&model.Server{})
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	query.Count(&total)
	err := query.Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&servers).Error
	return servers, total, err
}

func (r *ServerRepository) Update(server *model.Server) error {
	return r.db.Save(server).Error
}

func (r *ServerRepository) Delete(id uint) error {
	return r.db.Delete(&model.Server{}, id).Error
}

func (r *ServerRepository) CountByServerID(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Environment{}).Where("deploy_server_id = ?", id).Count(&count).Error
	return count, err
}
