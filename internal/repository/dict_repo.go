package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type DictRepository struct {
	db *gorm.DB
}

func NewDictRepository(db *gorm.DB) *DictRepository {
	return &DictRepository{db: db}
}

func (r *DictRepository) ListDictionaries() ([]model.Dictionary, error) {
	var dicts []model.Dictionary
	err := r.db.Order("id ASC").Find(&dicts).Error
	return dicts, err
}

func (r *DictRepository) FindDictionaryByID(id uint) (*model.Dictionary, error) {
	var dict model.Dictionary
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).First(&dict, id).Error
	return &dict, err
}

func (r *DictRepository) FindDictionaryByCode(code string) (*model.Dictionary, error) {
	var dict model.Dictionary
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).Where("code = ?", code).First(&dict).Error
	return &dict, err
}

func (r *DictRepository) CreateDictionary(dict *model.Dictionary) error {
	return r.db.Create(dict).Error
}

func (r *DictRepository) UpdateDictionary(dict *model.Dictionary) error {
	return r.db.Save(dict).Error
}

func (r *DictRepository) DeleteDictionary(id uint) error {
	return r.db.Delete(&model.Dictionary{}, id).Error
}

func (r *DictRepository) ListItemsByDictID(dictID uint) ([]model.DictItem, error) {
	var items []model.DictItem
	err := r.db.Where("dictionary_id = ?", dictID).Order("sort_order ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *DictRepository) CreateItem(item *model.DictItem) error {
	return r.db.Create(item).Error
}

func (r *DictRepository) FindItemByID(id uint) (*model.DictItem, error) {
	var item model.DictItem
	err := r.db.First(&item, id).Error
	return &item, err
}

func (r *DictRepository) UpdateItem(item *model.DictItem) error {
	return r.db.Save(item).Error
}

func (r *DictRepository) DeleteItem(id uint) error {
	return r.db.Delete(&model.DictItem{}, id).Error
}

func (r *DictRepository) DeleteItemsByDictID(dictID uint) error {
	return r.db.Where("dictionary_id = ?", dictID).Delete(&model.DictItem{}).Error
}

func (r *DictRepository) ReorderItems(dictID uint, itemIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range itemIDs {
			if err := tx.Model(&model.DictItem{}).Where("id = ? AND dictionary_id = ?", id, dictID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
