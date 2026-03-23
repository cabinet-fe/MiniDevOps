package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

func (r *CredentialRepository) Create(credential *model.Credential) error {
	return r.db.Create(credential).Error
}

func (r *CredentialRepository) Update(credential *model.Credential) error {
	return r.db.Save(credential).Error
}

func (r *CredentialRepository) Delete(id uint) error {
	return r.db.Delete(&model.Credential{}, id).Error
}

func (r *CredentialRepository) FindByID(id uint) (*model.Credential, error) {
	var credential model.Credential
	if err := r.db.First(&credential, id).Error; err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *CredentialRepository) FindByCreator(createdBy uint) ([]model.Credential, error) {
	var credentials []model.Credential
	if err := r.db.Where("created_by = ?", createdBy).Order("id DESC").Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *CredentialRepository) FindAll() ([]model.Credential, error) {
	var credentials []model.Credential
	if err := r.db.Order("id DESC").Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *CredentialRepository) FindByIDs(ids []uint) ([]model.Credential, error) {
	if len(ids) == 0 {
		return []model.Credential{}, nil
	}
	var credentials []model.Credential
	if err := r.db.Where("id IN ?", ids).Order("id DESC").Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}
