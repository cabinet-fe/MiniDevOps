package service

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// IService defines the interface for a generic CRUD service.
type IService[T any] interface {
	Create(c *fiber.Ctx, model *T) error
	GetByID(c *fiber.Ctx, id uint) (*T, error)
	GetAll(c *fiber.Ctx) ([]T, error)
	Update(c *fiber.Ctx, id uint, model *T) error
	Delete(c *fiber.Ctx, id uint) error
	GetPage(c *fiber.Ctx, page, pageSize int) ([]T, int64, error)
}

// CrudService provides a generic implementation for CRUD operations.
type CrudService[T any] struct {
	DB *gorm.DB
}

// NewCrudService creates a new instance of CrudService.
func NewCrudService[T any](db *gorm.DB) *CrudService[T] {
	return &CrudService[T]{DB: db}
}

// Create creates a new record in the database.
func (s *CrudService[T]) Create(c *fiber.Ctx, model *T) error {
	return s.DB.Create(model).Error
}

// GetByID retrieves a record by its ID.
func (s *CrudService[T]) GetByID(c *fiber.Ctx, id uint) (*T, error) {
	var model T
	if err := s.DB.First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// GetAll retrieves all records.
func (s *CrudService[T]) GetAll(c *fiber.Ctx) ([]T, error) {
	var models []T
	if err := s.DB.Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *CrudService[T]) GetPage(c *fiber.Ctx, page, pageSize int) ([]T, int64, error) {
	var models []T
	var total int64

	// 获取总数
	var zeroValue T
	if err := s.DB.Model(&zeroValue).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取分页数据
	if err := s.DB.Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return models, total, nil
}

// Update updates an existing record by its ID.
func (s *CrudService[T]) Update(c *fiber.Ctx, id uint, model *T) error {
	return s.DB.Model(new(T)).Where("id = ?", id).Updates(model).Error
}

// Delete removes a record by its ID.
func (s *CrudService[T]) Delete(c *fiber.Ctx, id uint) error {
	return s.DB.Delete(new(T), id).Error
}
