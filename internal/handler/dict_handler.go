package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type DictHandler struct {
	dictService *service.DictService
}

func NewDictHandler(ds *service.DictService) *DictHandler {
	return &DictHandler{dictService: ds}
}

func (h *DictHandler) ListDictionaries(c *gin.Context) {
	dicts, err := h.dictService.ListDictionaries()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, dicts)
}

func (h *DictHandler) CreateDictionary(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	dict := &model.Dictionary{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}
	if err := h.dictService.CreateDictionary(dict); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, dict)
}

func (h *DictHandler) GetDictionary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	dict, err := h.dictService.GetDictionary(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "字典不存在")
		return
	}
	pkg.Success(c, dict)
}

func (h *DictHandler) UpdateDictionary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	dict, err := h.dictService.GetDictionary(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "字典不存在")
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Code        *string `json:"code"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Name != nil {
		dict.Name = *req.Name
	}
	if req.Code != nil {
		dict.Code = *req.Code
	}
	if req.Description != nil {
		dict.Description = *req.Description
	}
	if err := h.dictService.UpdateDictionary(dict); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, dict)
}

func (h *DictHandler) DeleteDictionary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.dictService.DeleteDictionary(uint(id)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

func (h *DictHandler) ListItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	items, err := h.dictService.ListItems(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Success(c, items)
}

func (h *DictHandler) CreateItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	var req struct {
		Label     string `json:"label" binding:"required"`
		Value     string `json:"value" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Enabled   *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	item := &model.DictItem{
		DictionaryID: uint(id),
		Label:        req.Label,
		Value:        req.Value,
		SortOrder:    req.SortOrder,
		Enabled:      enabled,
	}
	if err := h.dictService.CreateItem(item); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, item)
}

func (h *DictHandler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	item, err := h.dictService.GetDictionary(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "字典不存在")
		return
	}
	_ = item

	existingItem, err := h.dictService.FindItemForUpdate(uint(itemID), uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "字典项不存在")
		return
	}
	var req struct {
		Label     *string `json:"label"`
		Value     *string `json:"value"`
		SortOrder *int    `json:"sort_order"`
		Enabled   *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Label != nil {
		existingItem.Label = *req.Label
	}
	if req.Value != nil {
		existingItem.Value = *req.Value
	}
	if req.SortOrder != nil {
		existingItem.SortOrder = *req.SortOrder
	}
	if req.Enabled != nil {
		existingItem.Enabled = *req.Enabled
	}
	if err := h.dictService.UpdateItem(existingItem); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, existingItem)
}

func (h *DictHandler) DeleteItem(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.dictService.DeleteItem(uint(itemID)); err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, nil)
}

func (h *DictHandler) GetItemsByCode(c *gin.Context) {
	code := c.Param("code")
	items, err := h.dictService.GetItemsByCode(code)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "字典不存在")
		return
	}
	pkg.Success(c, items)
}
