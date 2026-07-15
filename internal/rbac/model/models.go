package model

import "time"

const (
	ResourceTypeMenu   = "menu"
	ResourceTypePage   = "page"
	ResourceTypeAction = "action"
	ResourceTypeCard   = "card"
)

// Role is a customizable permission bundle.
type Role struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	Name        string           `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Code        string           `json:"code" gorm:"size:100;uniqueIndex;not null"`
	Description string           `json:"description" gorm:"size:500"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Permissions []RolePermission `json:"permissions,omitempty" gorm:"foreignKey:RoleID"`
}

func (Role) TableName() string { return "roles" }

// RolePermission binds a {path}:action code to a role.
type RolePermission struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	RoleID     uint   `json:"role_id" gorm:"index;not null"`
	Permission string `json:"permission" gorm:"size:200;not null;index"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// UserRole is the user↔role M2M join row.
type UserRole struct {
	UserID uint `json:"user_id" gorm:"primaryKey"`
	RoleID uint `json:"role_id" gorm:"primaryKey"`
}

func (UserRole) TableName() string { return "user_roles" }

// RbacResource is the unique permission/menu resource tree node.
type RbacResource struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Path         string         `json:"path" gorm:"size:200;uniqueIndex;not null"`
	Type         string         `json:"type" gorm:"size:20;not null;index"`
	ParentID     *uint          `json:"parent_id" gorm:"index"`
	Enabled      bool           `json:"enabled" gorm:"not null;default:true"`
	SortKey      int            `json:"sort_key" gorm:"not null;default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	MenuMetadata *MenuMetadata  `json:"menu_metadata,omitempty" gorm:"foreignKey:ResourceID"`
	Children     []RbacResource `json:"children,omitempty" gorm:"-"`
}

func (RbacResource) TableName() string { return "rbac_resources" }

// MenuMetadata is 1:1 with a menu-type RbacResource.
type MenuMetadata struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	ResourceID uint   `json:"resource_id" gorm:"uniqueIndex;not null"`
	Title      string `json:"title" gorm:"size:100;not null"`
	Route      string `json:"route" gorm:"size:200"`
	IconBase64 string `json:"icon_base64,omitempty" gorm:"type:text"`
	IconMime   string `json:"icon_mime,omitempty" gorm:"size:64"`
}

func (MenuMetadata) TableName() string { return "menu_metadata" }

// MenuNode is the trimmed menu tree returned by login /auth/me.
// Frontend maps title/route/icon/children → @veltra/desktop NavItem (path ← route).
type MenuNode struct {
	Path     string     `json:"path"`
	Title    string     `json:"title"`
	Route    string     `json:"route,omitempty"`
	Icon     string     `json:"icon,omitempty"`
	Sort     int        `json:"sort"`
	Children []MenuNode `json:"children,omitempty"`
}
