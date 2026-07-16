package repository

import (
	"time"

	"bedrock/internal/storage/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StorageRepository struct {
	db *gorm.DB
}

func NewStorageRepository(db *gorm.DB) *StorageRepository {
	return &StorageRepository{db: db}
}

func (r *StorageRepository) Create(object *model.StorageObject) error {
	return r.db.Create(object).Error
}

func (r *StorageRepository) FindByID(id uint) (*model.StorageObject, error) {
	var object model.StorageObject
	if err := r.db.First(&object, id).Error; err != nil {
		return nil, err
	}
	return &object, nil
}

func (r *StorageRepository) FindBySHA256(sha256 string) (*model.StorageObject, error) {
	var object model.StorageObject
	if err := r.db.Where("sha256 = ?", sha256).First(&object).Error; err != nil {
		return nil, err
	}
	return &object, nil
}

func (r *StorageRepository) FindBySHA256IncludingDeleted(sha256 string) (*model.StorageObject, error) {
	var object model.StorageObject
	if err := r.db.Unscoped().Where("sha256 = ?", sha256).First(&object).Error; err != nil {
		return nil, err
	}
	return &object, nil
}

func (r *StorageRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&model.StorageObject{}).Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": nil, "purge_after": nil, "ref_count": 1}).Error
}

func (r *StorageRepository) IncrementRef(id uint) error {
	return r.db.Model(&model.StorageObject{}).Where("id = ?", id).
		UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error
}

// DecrementRef atomically records a released reference and returns the object
// after the decrement. Callers delete the backing file only when RefCount is 0.
func (r *StorageRepository) DecrementRef(id uint, purgeAfter *time.Time) (*model.StorageObject, error) {
	var object model.StorageObject
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&object, id).Error; err != nil {
			return err
		}
		if object.RefCount > 0 {
			object.RefCount--
		}
		if object.RefCount == 0 {
			object.PurgeAfter = purgeAfter
		}
		if err := tx.Save(&object).Error; err != nil {
			return err
		}
		if object.RefCount == 0 {
			return tx.Delete(&object).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &object, nil
}

func (r *StorageRepository) Delete(id uint) error {
	return r.db.Delete(&model.StorageObject{}, id).Error
}

func (r *StorageRepository) Purge(id uint) error {
	return r.db.Unscoped().Delete(&model.StorageObject{}, id).Error
}
