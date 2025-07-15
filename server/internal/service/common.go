package service

import (
	"reflect"
	"server/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CrudService provides a generic implementation for CRUD operations.
type CrudService[T any] struct {
	DB    *gorm.DB
	Model *T
}

// NewCrudService creates a new instance of CrudService.
func NewCrudService[M any](db *gorm.DB) *CrudService[M] {
	return &CrudService[M]{db, new(M)}
}

var typeQueryMap = map[string]func(string) (string, string){
	"string": func(value string) (string, string) {
		return "a LIKE ?", "%" + value + "%"
	},
	"int": func(value string) (string, string) {
		return "a = ?", value
	},
	"float": func(value string) (string, string) {
		return "a = ?", value
	},
	"bool": func(value string) (string, string) {
		return "a = ?", value
	},
	"time": func(value string) (string, string) {
		return "a = ?", value
	},
}

func (s *CrudService[T]) buildQuery(c *fiber.Ctx) *gorm.DB {
	query := s.DB

	modelType := reflect.TypeOf(s.Model).Elem()
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		q, v := typeQueryMap[field.Type.String()](field.Name)
		query = query.Where(q, v)
	}
	return query
}

// 新增一条记录
func (s *CrudService[T]) Create(c *fiber.Ctx) error {
	var model T
	if err := c.BodyParser(s.Model); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err)
	}
	return s.DB.Create(&model).Error
}

// 根据ID获取一条记录
func (s *CrudService[T]) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err, "ID 参数错误")
	}

	var model T
	if err := s.DB.First(&model, id).Error; err != nil {
		return utils.Error(c, fiber.StatusNotFound, err)
	}
	return utils.SuccessWithData(c, &model)
}

// 获取列表（保持向后兼容）
func (s *CrudService[T]) GetList(c *fiber.Ctx) error {
	var models []T

	query := s.buildQuery(c)

	if err := query.Find(&models).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err)
	}
	return utils.SuccessWithData(c, &models)
}

// 获取分页数据（保持向后兼容）
func (s *CrudService[T]) GetPage(c *fiber.Ctx) error {
	queries := c.Queries()
	var models []T
	var total int64

	page, err := strconv.Atoi(queries["page"])
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err)
	}
	pageSize, err := strconv.Atoi(queries["page_size"])
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err)
	}

	if err := s.DB.Model(&s.Model).Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err)
	}

	offset := (page - 1) * pageSize
	if err := s.DB.Model(&s.Model).Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err)
	}

	return utils.SuccessWithData(c, &models)
}

// Update updates an existing record by its ID.
func (s *CrudService[T]) Update(c *fiber.Ctx, id uint, model *T) error {
	return s.DB.Model(new(T)).Where("id = ?", id).Updates(model).Error
}

// Delete removes a record by its ID.
func (s *CrudService[T]) Delete(c *fiber.Ctx, id uint) error {
	return s.DB.Delete(new(T), id).Error
}

// GetModelByID 通用方法，用于其他服务调用
func (s *CrudService[T]) GetModelByID(id uint) (*T, error) {
	var model T
	if err := s.DB.First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}
