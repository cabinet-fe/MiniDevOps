package repository

import (
	"buildflow/internal/model"

	"gorm.io/gorm"
)

type VarGroupRepository struct {
	db *gorm.DB
}

func NewVarGroupRepository(db *gorm.DB) *VarGroupRepository {
	return &VarGroupRepository{db: db}
}

func (r *VarGroupRepository) List() ([]model.VarGroup, error) {
	var groups []model.VarGroup
	err := r.db.Preload("Items").Order("name ASC").Find(&groups).Error
	return groups, err
}

func (r *VarGroupRepository) FindByID(id uint) (*model.VarGroup, error) {
	var group model.VarGroup
	err := r.db.Preload("Items").First(&group, id).Error
	return &group, err
}

func (r *VarGroupRepository) FindByName(name string) (*model.VarGroup, error) {
	var group model.VarGroup
	err := r.db.Preload("Items").Where("name = ?", name).First(&group).Error
	return &group, err
}

func (r *VarGroupRepository) Create(group *model.VarGroup) error {
	return r.db.Create(group).Error
}

func (r *VarGroupRepository) ReplaceItems(groupID uint, items []model.VarGroupItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("var_group_id = ?", groupID).Delete(&model.VarGroupItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].VarGroupID = groupID
		}
		return tx.Create(&items).Error
	})
}

func (r *VarGroupRepository) Update(group *model.VarGroup) error {
	return r.db.Save(group).Error
}

func (r *VarGroupRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("var_group_id = ?", id).Delete(&model.EnvironmentVarGroup{}).Error; err != nil {
			return err
		}
		if err := tx.Where("var_group_id = ?", id).Delete(&model.VarGroupItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.VarGroup{}, id).Error
	})
}

func (r *VarGroupRepository) ListItemsByEnvironmentID(environmentID uint) ([]model.VarGroupItem, error) {
	var items []model.VarGroupItem
	err := r.db.
		Joins("JOIN environment_var_groups ON environment_var_groups.var_group_id = var_group_items.var_group_id").
		Where("environment_var_groups.environment_id = ?", environmentID).
		Order("var_group_items.id ASC").
		Find(&items).Error
	return items, err
}

func (r *VarGroupRepository) ListEnvironmentVarGroupIDs(environmentID uint) ([]uint, error) {
	var links []model.EnvironmentVarGroup
	if err := r.db.Where("environment_id = ?", environmentID).Order("var_group_id ASC").Find(&links).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.VarGroupID)
	}
	return ids, nil
}

func (r *VarGroupRepository) SetEnvironmentVarGroupIDs(environmentID uint, groupIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("environment_id = ?", environmentID).Delete(&model.EnvironmentVarGroup{}).Error; err != nil {
			return err
		}
		if len(groupIDs) == 0 {
			return nil
		}
		links := make([]model.EnvironmentVarGroup, 0, len(groupIDs))
		for _, groupID := range groupIDs {
			links = append(links, model.EnvironmentVarGroup{
				EnvironmentID: environmentID,
				VarGroupID:    groupID,
			})
		}
		return tx.Create(&links).Error
	})
}

func (r *VarGroupRepository) DeleteEnvironmentLinks(environmentID uint) error {
	return r.db.Where("environment_id = ?", environmentID).Delete(&model.EnvironmentVarGroup{}).Error
}

func (r *VarGroupRepository) DeleteLinksByProjectID(projectID uint) error {
	return r.db.Where(
		"environment_id IN (?)",
		r.db.Model(&model.Environment{}).Select("id").Where("project_id = ?", projectID),
	).Delete(&model.EnvironmentVarGroup{}).Error
}
