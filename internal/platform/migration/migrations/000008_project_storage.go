package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000008_project_storage", upProjectStorage)
}

func upProjectStorage(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	models := []interface{}{
		&storageObjectMigrationModel{},
		&productProjectMigrationModel{},
		&projectMemberMigrationModel{},
		&requirementMigrationModel{},
		&requirementCommentMigrationModel{},
		&requirementAttachmentMigrationModel{},
		&apiDocNodeMigrationModel{},
	}
	for _, item := range models {
		if db.Migrator().HasTable(item) {
			continue
		}
		if err := db.Migrator().CreateTable(item); err != nil {
			return err
		}
	}
	return seedRequirementStatuses(db)
}

type storageObjectMigrationModel struct {
	ID          uint           `gorm:"primaryKey"`
	Kind        string         `gorm:"size:32;not null;index"`
	SHA256      string         `gorm:"size:64;not null;uniqueIndex"`
	Size        int64          `gorm:"not null"`
	ContentType string         `gorm:"size:200"`
	Path        string         `gorm:"size:500;not null"`
	RefCount    int            `gorm:"not null;default:1"`
	CreatedBy   uint           `gorm:"index"`
	PurgeAfter  *time.Time     `gorm:""`
	CreatedAt   time.Time      `gorm:""`
	UpdatedAt   time.Time      `gorm:""`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (storageObjectMigrationModel) TableName() string { return "storage_objects" }

type productProjectMigrationModel struct {
	ID           uint           `gorm:"primaryKey"`
	Name         string         `gorm:"size:200;not null"`
	Slug         string         `gorm:"size:120;not null;uniqueIndex"`
	Description  string         `gorm:"type:text"`
	Status       string         `gorm:"size:20;not null;default:active;index"`
	OwnerID      uint           `gorm:"not null;index"`
	RepositoryID *uint          `gorm:"index"`
	Tags         string         `gorm:""`
	CreatedBy    uint           `gorm:"index"`
	CreatedAt    time.Time      `gorm:""`
	UpdatedAt    time.Time      `gorm:""`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (productProjectMigrationModel) TableName() string { return "product_projects" }

type projectMemberMigrationModel struct {
	ID        uint      `gorm:"primaryKey"`
	ProjectID uint      `gorm:"not null;uniqueIndex:idx_project_member"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_project_member;index"`
	Role      string    `gorm:"size:20;not null;index"`
	CreatedAt time.Time `gorm:""`
	UpdatedAt time.Time `gorm:""`
}

func (projectMemberMigrationModel) TableName() string { return "project_members" }

type requirementMigrationModel struct {
	ID           uint           `gorm:"primaryKey"`
	ProjectID    uint           `gorm:"not null;index"`
	Title        string         `gorm:"size:500;not null"`
	Description  string         `gorm:"type:text"`
	Status       string         `gorm:"size:100;not null;index"`
	Priority     string         `gorm:"size:30;not null;default:normal;index"`
	AssigneeID   *uint          `gorm:"index"`
	RepositoryID *uint          `gorm:"index"`
	Tags         string         `gorm:""`
	CreatedBy    uint           `gorm:"index"`
	UpdatedBy    uint           `gorm:"index"`
	CreatedAt    time.Time      `gorm:""`
	UpdatedAt    time.Time      `gorm:""`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (requirementMigrationModel) TableName() string { return "requirements" }

type requirementCommentMigrationModel struct {
	ID            uint           `gorm:"primaryKey"`
	RequirementID uint           `gorm:"not null;index"`
	Content       string         `gorm:"type:text;not null"`
	CreatedBy     uint           `gorm:"index"`
	CreatedAt     time.Time      `gorm:""`
	UpdatedAt     time.Time      `gorm:""`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (requirementCommentMigrationModel) TableName() string { return "requirement_comments" }

type requirementAttachmentMigrationModel struct {
	ID              uint      `gorm:"primaryKey"`
	RequirementID   uint      `gorm:"not null;index"`
	StorageObjectID uint      `gorm:"not null;uniqueIndex"`
	Filename        string    `gorm:"size:500;not null"`
	CreatedBy       uint      `gorm:"index"`
	CreatedAt       time.Time `gorm:""`
}

func (requirementAttachmentMigrationModel) TableName() string { return "requirement_attachments" }

type apiDocNodeMigrationModel struct {
	ID               uint           `gorm:"primaryKey"`
	ProjectID        uint           `gorm:"not null;index"`
	ParentID         *uint          `gorm:"index"`
	Kind             string         `gorm:"size:10;not null;index"`
	Name             string         `gorm:"size:300;not null"`
	SortOrder        int            `gorm:"not null;default:0;index"`
	RepositoryID     *uint          `gorm:"index"`
	PublishedContent string         `gorm:"type:text"`
	DraftContent     string         `gorm:"type:text"`
	ContentVersion   int            `gorm:"not null;default:0"`
	DraftBaseVersion int            `gorm:"not null;default:0"`
	DraftUpdatedAt   *time.Time     `gorm:""`
	DraftSourceRunID *uint          `gorm:"index"`
	CreatedBy        uint           `gorm:"index"`
	UpdatedBy        uint           `gorm:"index"`
	CreatedAt        time.Time      `gorm:""`
	UpdatedAt        time.Time      `gorm:""`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (apiDocNodeMigrationModel) TableName() string { return "api_doc_nodes" }

type requirementStatusDictionaryMigrationModel struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Code        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"size:500"`
	CreatedAt   time.Time `gorm:""`
	UpdatedAt   time.Time `gorm:""`
}

func (requirementStatusDictionaryMigrationModel) TableName() string { return "dictionaries" }

type requirementStatusItemMigrationModel struct {
	ID           uint      `gorm:"primaryKey"`
	DictionaryID uint      `gorm:"index;not null"`
	Label        string    `gorm:"size:200;not null"`
	Value        string    `gorm:"size:200;not null"`
	SortOrder    int       `gorm:"not null;default:0"`
	Enabled      bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:""`
	UpdatedAt    time.Time `gorm:""`
}

func (requirementStatusItemMigrationModel) TableName() string { return "dict_items" }

func seedRequirementStatuses(db *gorm.DB) error {
	dictionary := requirementStatusDictionaryMigrationModel{
		Name: "需求状态", Code: "requirement_status", Description: "产品项目需求的可扩展状态字典",
	}
	if err := db.Where("code = ?", dictionary.Code).FirstOrCreate(&dictionary).Error; err != nil {
		return err
	}
	items := []requirementStatusItemMigrationModel{
		{DictionaryID: dictionary.ID, Label: "待梳理", Value: "backlog", SortOrder: 10, Enabled: true},
		{DictionaryID: dictionary.ID, Label: "待处理", Value: "todo", SortOrder: 20, Enabled: true},
		{DictionaryID: dictionary.ID, Label: "进行中", Value: "doing", SortOrder: 30, Enabled: true},
		{DictionaryID: dictionary.ID, Label: "已完成", Value: "done", SortOrder: 40, Enabled: true},
		{DictionaryID: dictionary.ID, Label: "已取消", Value: "cancelled", SortOrder: 50, Enabled: true},
	}
	for _, item := range items {
		if err := db.Where("dictionary_id = ? AND value = ?", item.DictionaryID, item.Value).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}
	return nil
}
