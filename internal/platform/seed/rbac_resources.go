package seed

import (
	"errors"
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

// EnsureRBACResources seeds and incrementally completes the preset resource tree.
// Ops paths are included but only effective for super-admin. Incremental upserts
// keep new cards available on installations that were initialized before P2.
func EnsureRBACResources(db *gorm.DB) error {
	tree := []seedMenu{
		{Path: "dashboard", Title: "仪表盘", Route: "/", SortKey: 10},
		{
			Path: "ops", Title: "运维", Route: "/ops", SortKey: 20,
			Children: []seedMenu{
				{Path: "ops.processes", Title: "进程", Route: "/ops/processes", SortKey: 10},
				{Path: "ops.dev_environments", Title: "开发环境", Route: "/ops/dev-environments", SortKey: 20},
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
			Path: "project", Title: "项目管理", Route: "/project", SortKey: 40,
			Children: []seedMenu{
				{Path: "project.projects", Title: "产品项目", Route: "/project/projects", SortKey: 10},
				{Path: "project.requirements", Title: "需求管理", Route: "/project/requirements", SortKey: 20},
				{Path: "project.docs", Title: "接口文档", Route: "/project/docs", SortKey: 30},
			},
		},
		{
			Path: "ai", Title: "AI", Route: "/ai", SortKey: 50,
			Children: []seedMenu{
				{Path: "ai.clis", Title: "AI CLI", Route: "/ai/clis", SortKey: 10},
				{Path: "ai.agents", Title: "智能体", Route: "/ai/agents", SortKey: 20},
				{Path: "ai.skills", Title: "Skills", Route: "/ai/skills", SortKey: 30},
				{Path: "ai.tokens", Title: "访问令牌", Route: "/ai/tokens", SortKey: 40},
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
		if err := seedDashboardCards(tx, now); err != nil {
			return err
		}
		if err := seedProjectScopeActions(tx, now); err != nil {
			return err
		}
		return removeRetiredMenu(tx, "ops.toolchains")
	})
}

func insertMenu(tx *gorm.DB, node seedMenu, parentID *uint, now time.Time) error {
	var res model.RbacResource
	err := tx.Where("path = ?", node.Path).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		res = model.RbacResource{
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
	} else if err != nil {
		return fmt.Errorf("find resource %s: %w", node.Path, err)
	} else if res.SortKey != node.SortKey {
		if err := tx.Model(&res).Updates(map[string]any{
			"sort_key":   node.SortKey,
			"updated_at": now,
		}).Error; err != nil {
			return fmt.Errorf("update resource %s: %w", node.Path, err)
		}
	}
	var meta model.MenuMetadata
	err = tx.Where("resource_id = ?", res.ID).First(&meta).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		meta = model.MenuMetadata{
			ResourceID: res.ID,
			Title:      node.Title,
			Route:      node.Route,
		}
		if err := tx.Create(&meta).Error; err != nil {
			return fmt.Errorf("create menu metadata %s: %w", node.Path, err)
		}
	} else if err != nil {
		return fmt.Errorf("find menu metadata %s: %w", node.Path, err)
	} else if meta.Title != node.Title || meta.Route != node.Route {
		// Keep preset routes/titles in sync when seeds evolve (e.g. P3 path fix).
		if err := tx.Model(&meta).Updates(map[string]any{
			"title": node.Title,
			"route": node.Route,
		}).Error; err != nil {
			return fmt.Errorf("update menu metadata %s: %w", node.Path, err)
		}
	}
	id := res.ID
	for _, child := range node.Children {
		if err := insertMenu(tx, child, &id, now); err != nil {
			return err
		}
	}
	return nil
}

func seedDashboardCards(tx *gorm.DB, now time.Time) error {
	var dashboard model.RbacResource
	if err := tx.Where("path = ?", "dashboard").First(&dashboard).Error; err != nil {
		return err
	}
	for index, path := range []string{
		"dashboard.build_summary",
		"dashboard.system_info",
		"dashboard.system_status",
	} {
		var existing model.RbacResource
		err := tx.Where("path = ?", path).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			parentID := dashboard.ID
			card := model.RbacResource{
				Path: path, Type: model.ResourceTypeCard, ParentID: &parentID,
				Enabled: true, SortKey: (index + 1) * 10, CreatedAt: now, UpdatedAt: now,
			}
			if err := tx.Create(&card).Error; err != nil {
				return fmt.Errorf("create dashboard card %s: %w", path, err)
			}
		} else if err != nil {
			return fmt.Errorf("find dashboard card %s: %w", path, err)
		}
	}
	return nil
}

// seedProjectScopeActions makes the project-wide ACL bypasses visible in the
// resource tree. Their paths are complete permission codes because actions are
// assigned to roles as {resource path}:{action}, while ordinary menu resources
// represent only the resource path.
func seedProjectScopeActions(tx *gorm.DB, now time.Time) error {
	var projects model.RbacResource
	if err := tx.Where("path = ?", "project.projects").First(&projects).Error; err != nil {
		return fmt.Errorf("find project projects resource: %w", err)
	}

	for index, permission := range []string{
		"project.projects:view_all",
		"project.projects:manage_all",
	} {
		var existing model.RbacResource
		err := tx.Where("path = ?", permission).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			parentID := projects.ID
			action := model.RbacResource{
				Path:      permission,
				Type:      model.ResourceTypeAction,
				ParentID:  &parentID,
				Enabled:   true,
				SortKey:   100 + (index+1)*10,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(&action).Error; err != nil {
				return fmt.Errorf("create project scope action %s: %w", permission, err)
			}
		} else if err != nil {
			return fmt.Errorf("find project scope action %s: %w", permission, err)
		}
	}
	return nil
}

func removeRetiredMenu(tx *gorm.DB, path string) error {
	var res model.RbacResource
	err := tx.Where("path = ?", path).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find retired resource %s: %w", path, err)
	}
	if err := tx.Where("resource_id = ?", res.ID).Delete(&model.MenuMetadata{}).Error; err != nil {
		return fmt.Errorf("delete menu metadata %s: %w", path, err)
	}
	if err := tx.Where("permission LIKE ?", path+":%").Delete(&model.RolePermission{}).Error; err != nil {
		return fmt.Errorf("delete role permissions %s: %w", path, err)
	}
	if err := tx.Delete(&res).Error; err != nil {
		return fmt.Errorf("delete retired resource %s: %w", path, err)
	}
	return nil
}
