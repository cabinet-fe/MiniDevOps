package service

import (
	"fmt"

	"buildflow/internal/model"
	"buildflow/internal/repository"
)

type DictService struct {
	repo *repository.DictRepository
}

func NewDictService(repo *repository.DictRepository) *DictService {
	return &DictService{repo: repo}
}

func (s *DictService) ListDictionaries() ([]model.Dictionary, error) {
	return s.repo.ListDictionaries()
}

func (s *DictService) GetDictionary(id uint) (*model.Dictionary, error) {
	return s.repo.FindDictionaryByID(id)
}

func (s *DictService) CreateDictionary(dict *model.Dictionary) error {
	return s.repo.CreateDictionary(dict)
}

func (s *DictService) UpdateDictionary(dict *model.Dictionary) error {
	return s.repo.UpdateDictionary(dict)
}

func (s *DictService) DeleteDictionary(id uint) error {
	if err := s.repo.DeleteItemsByDictID(id); err != nil {
		return err
	}
	return s.repo.DeleteDictionary(id)
}

func (s *DictService) ListItems(dictID uint) ([]model.DictItem, error) {
	return s.repo.ListItemsByDictID(dictID)
}

func (s *DictService) GetItemsByCode(code string) ([]model.DictItem, error) {
	dict, err := s.repo.FindDictionaryByCode(code)
	if err != nil {
		return nil, err
	}
	var enabled []model.DictItem
	for _, item := range dict.Items {
		if item.Enabled {
			enabled = append(enabled, item)
		}
	}
	return enabled, nil
}

func (s *DictService) FindItemForUpdate(itemID, dictID uint) (*model.DictItem, error) {
	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return nil, err
	}
	if item.DictionaryID != dictID {
		return nil, fmt.Errorf("字典项不属于该字典")
	}
	return item, nil
}

func (s *DictService) CreateItem(item *model.DictItem) error {
	return s.repo.CreateItem(item)
}

func (s *DictService) UpdateItem(item *model.DictItem) error {
	return s.repo.UpdateItem(item)
}

func (s *DictService) DeleteItem(id uint) error {
	return s.repo.DeleteItem(id)
}

func (s *DictService) ReorderItems(dictID uint, itemIDs []uint) error {
	return s.repo.ReorderItems(dictID, itemIDs)
}
