package seed

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/rbac/model"
)

type seedMenu struct {
	Path     string
	Title    string
	Route    string
	SortKey  int
	Children []seedMenu
}

// EnsureRBACResources seeds the preset menu/resource tree for system.* and cicd.*
// (ops paths are included but only effective for super-admin).
func EnsureRBACResources(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.RbacResource{}).Count(&count).Error; err != nil {
		return fmt.Errorf("counting rbac resources: %w", err)
	}
	if count > 0 {
		return nil
	}

	tree := []seedMenu{
		{Path: "dashboard", Title: "仪表盘", Route: "/", SortKey: 10},
		{
			Path: "ops", Title: "运维", Route: "/ops", SortKey: 20,
			Children: []seedMenu{
				{Path: "ops.processes", Title: "进程", Route: "/ops/processes", SortKey: 10},
				{Path: "ops.toolchains", Title: "工具链", Route: "/ops/toolchains", SortKey: 20},
			},
		},
		{
			Path: "cicd", Title: "CI/CD", Route: "/cicd", SortKey: 30,
			Children: []seedMenu{
				{Path: "cicd.repositories", Title: "代码仓库", Route: "/cicd/repositories", SortKey: 10},
				{Path: "cicd.build_jobs", Title: "构建任务", Route: "/cicd/build-jobs", SortKey: 20},
				{Path: "cicd.build_runs", Title: "构建执行", Route: "/cicd/build-runs", SortKey: 30},
				{Path: "cicd.servers", Title: "服务器", Route: "/cicd/servers", SortKey: 40},
				{Path: "cicd.credentials", Title: "凭证", Route: "/cicd/credentials", SortKey: 50},
			},
		},
		{
			Path: "system", Title: "系统管理", Route: "/system", SortKey: 90,
			Children: []seedMenu{
				{Path: "system.users", Title: "用户", Route: "/system/users", SortKey: 10},
				{Path: "system.roles", Title: "角色", Route: "/system/roles", SortKey: 20},
				{Path: "system.resources", Title: "权限资源", Route: "/system/resources", SortKey: 30},
				{Path: "system.dictionaries", Title: "字典", Route: "/system/dictionaries", SortKey: 40},
				{Path: "system.operation_logs", Title: "操作日志", Route: "/system/operation-logs", SortKey: 50},
			},
		},
	}

	now := time.Now().UTC()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, root := range tree {
			if err := insertMenu(tx, root, nil, now); err != nil {
				return err
			}
		}
		return nil
	})
}

func insertMenu(tx *gorm.DB, node seedMenu, parentID *uint, now time.Time) error {
	res := model.RbacResource{
		Path:      node.Path,
		Type:      model.ResourceTypeMenu,
		ParentID:  parentID,
		Enabled:   true,
		SortKey:   node.SortKey,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := tx.Create(&res).Error; err != nil {
		return fmt.Errorf("create resource %s: %w", node.Path, err)
	}
	meta := model.MenuMetadata{
		ResourceID: res.ID,
		Title:      node.Title,
		Route:      node.Route,
	}
	if err := tx.Create(&meta).Error; err != nil {
		return fmt.Errorf("create menu metadata %s: %w", node.Path, err)
	}
	id := res.ID
	for _, child := range node.Children {
		if err := insertMenu(tx, child, &id, now); err != nil {
			return err
		}
	}
	return nil
}
