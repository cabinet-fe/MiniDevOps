package repository

import (
	"bedrock/internal/rbac/model"

	"gorm.io/gorm"
)

type ResourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Create(res *model.RbacResource) error {
	return r.db.Create(res).Error
}

func (r *ResourceRepository) FindByID(id uint) (*model.RbacResource, error) {
	var res model.RbacResource
	err := r.db.Preload("MenuMetadata").First(&res, id).Error
	return &res, err
}

func (r *ResourceRepository) ListAll() ([]model.RbacResource, error) {
	var items []model.RbacResource
	err := r.db.Preload("MenuMetadata").Order("sort_key ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *ResourceRepository) ListMenus() ([]model.RbacResource, error) {
	var items []model.RbacResource
	err := r.db.Preload("MenuMetadata").
		Where("type = ?", model.ResourceTypeMenu).
		Order("sort_key ASC, id ASC").
		Find(&items).Error
	return items, err
}

func (r *ResourceRepository) Update(res *model.RbacResource) error {
	return r.db.Save(res).Error
}

func (r *ResourceRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("resource_id = ?", id).Delete(&model.MenuMetadata{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.RbacResource{}, id).Error
	})
}

func (r *ResourceRepository) CountChildren(parentID uint) (int64, error) {
	var n int64
	err := r.db.Model(&model.RbacResource{}).Where("parent_id = ?", parentID).Count(&n).Error
	return n, err
}

func (r *ResourceRepository) UpsertMenuMetadata(meta *model.MenuMetadata) error {
	var existing model.MenuMetadata
	err := r.db.Where("resource_id = ?", meta.ResourceID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(meta).Error
	}
	if err != nil {
		return err
	}
	meta.ID = existing.ID
	return r.db.Save(meta).Error
}
