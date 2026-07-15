package service

import (
	"sort"
	"strings"

	"bedrock/internal/rbac"
	"bedrock/internal/rbac/model"
	"bedrock/internal/rbac/repository"
)

// PermissionService computes permission unions and trimmed menu trees.
type PermissionService struct {
	roles     *repository.RoleRepository
	resources *repository.ResourceRepository
}

func NewPermissionService(roles *repository.RoleRepository, resources *repository.ResourceRepository) *PermissionService {
	return &PermissionService{roles: roles, resources: resources}
}

// ResolvePermissions returns the effective permission code set for a user.
// Super-admin receives every known resource action derived from the resource tree
// (including ops). Non-super-admin gets role union with ops codes stripped.
func (s *PermissionService) ResolvePermissions(userID uint, isSuperAdmin bool) ([]string, error) {
	if isSuperAdmin {
		return s.allPermissions()
	}
	codes, err := s.roles.ListPermissionsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return rbac.FilterOpsPermissions(uniqSorted(codes)), nil
}

// CheckAccess returns nil if the user may perform required permission.
// Ops paths always require super-admin regardless of role grants.
func (s *PermissionService) CheckAccess(userID uint, isSuperAdmin bool, required string) error {
	if required == "" {
		return errForbidden("missing permission")
	}
	if rbac.IsOpsPermission(required) && !isSuperAdmin {
		return errForbidden("仅超级管理员可访问运维功能")
	}
	if isSuperAdmin {
		return nil
	}
	codes, err := s.ResolvePermissions(userID, false)
	if err != nil {
		return err
	}
	if !rbac.HasPermission(rbac.ToSet(codes), required) {
		return errForbidden("没有权限: " + required)
	}
	return nil
}

// TrimMenus builds the menu tree for /auth/me.
// Rules: leaf needs own :view; parents auto-filled when a visible descendant exists;
// ops menus only for super-admin; disabled resources omitted.
func (s *PermissionService) TrimMenus(userID uint, isSuperAdmin bool) ([]model.MenuNode, error) {
	menus, err := s.resources.ListMenus()
	if err != nil {
		return nil, err
	}
	var viewSet map[string]struct{}
	if isSuperAdmin {
		viewSet = map[string]struct{}{}
		for _, m := range menus {
			viewSet[m.Path] = struct{}{}
		}
	} else {
		perms, err := s.ResolvePermissions(userID, false)
		if err != nil {
			return nil, err
		}
		viewSet = map[string]struct{}{}
		for _, p := range perms {
			path, action, ok := rbac.SplitPermission(p)
			if ok && action == "view" {
				viewSet[path] = struct{}{}
			}
		}
	}
	return buildTrimmedMenuTree(menus, viewSet, isSuperAdmin), nil
}

var commonActions = []string{
	"view", "create", "update", "delete", "execute",
	"download", "cancel", "retry", "redeploy", "install", "test", "use",
	"view_all", "manage_all",
}

func (s *PermissionService) allPermissions() ([]string, error) {
	resources, err := s.resources.ListAll()
	if err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, res := range resources {
		for _, action := range commonActions {
			set[rbac.PermissionCode(res.Path, action)] = struct{}{}
		}
	}
	stored, err := s.roles.ListDistinctPermissions()
	if err != nil {
		return nil, err
	}
	for _, p := range stored {
		set[p] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

type forbiddenError struct{ msg string }

func (e *forbiddenError) Error() string { return e.msg }

func errForbidden(msg string) error { return &forbiddenError{msg: msg} }

// IsForbidden reports whether err is a permission denial.
func IsForbidden(err error) bool {
	_, ok := err.(*forbiddenError)
	return ok
}

func uniqSorted(codes []string) []string {
	set := rbac.ToSet(codes)
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

type menuFlat struct {
	res  model.RbacResource
	meta *model.MenuMetadata
}

func buildTrimmedMenuTree(menus []model.RbacResource, viewSet map[string]struct{}, isSuperAdmin bool) []model.MenuNode {
	byID := make(map[uint]*menuFlat, len(menus))
	childrenOf := map[uint][]uint{}
	var roots []uint

	for i := range menus {
		m := &menus[i]
		if !m.Enabled {
			continue
		}
		if rbac.IsOpsPath(m.Path) && !isSuperAdmin {
			continue
		}
		byID[m.ID] = &menuFlat{res: *m, meta: m.MenuMetadata}
		if m.ParentID == nil {
			roots = append(roots, m.ID)
		} else {
			childrenOf[*m.ParentID] = append(childrenOf[*m.ParentID], m.ID)
		}
	}

	var build func(id uint) (model.MenuNode, bool)
	build = func(id uint) (model.MenuNode, bool) {
		flat, ok := byID[id]
		if !ok {
			return model.MenuNode{}, false
		}
		childIDs := childrenOf[id]
		sort.Slice(childIDs, func(i, j int) bool {
			a, b := byID[childIDs[i]], byID[childIDs[j]]
			if a.res.SortKey != b.res.SortKey {
				return a.res.SortKey < b.res.SortKey
			}
			return a.res.ID < b.res.ID
		})

		var kids []model.MenuNode
		for _, cid := range childIDs {
			if node, ok := build(cid); ok {
				kids = append(kids, node)
			}
		}

		_, hasView := viewSet[flat.res.Path]
		isLeaf := len(childIDs) == 0
		include := false
		if isLeaf {
			include = hasView
		} else {
			// Parent auto-filled when any visible descendant exists.
			include = len(kids) > 0
		}
		if !include {
			return model.MenuNode{}, false
		}

		title := flat.res.Path
		route := ""
		icon := ""
		if flat.meta != nil {
			if flat.meta.Title != "" {
				title = flat.meta.Title
			}
			route = flat.meta.Route
			if flat.meta.IconBase64 != "" {
				if flat.meta.IconMime != "" && !strings.HasPrefix(flat.meta.IconBase64, "data:") {
					icon = "data:" + flat.meta.IconMime + ";base64," + flat.meta.IconBase64
				} else {
					icon = flat.meta.IconBase64
				}
			}
		}
		return model.MenuNode{
			Path:     flat.res.Path,
			Title:    title,
			Route:    route,
			Icon:     icon,
			Sort:     flat.res.SortKey,
			Children: kids,
		}, true
	}

	sort.Slice(roots, func(i, j int) bool {
		a, b := byID[roots[i]], byID[roots[j]]
		if a.res.SortKey != b.res.SortKey {
			return a.res.SortKey < b.res.SortKey
		}
		return a.res.ID < b.res.ID
	})

	out := make([]model.MenuNode, 0, len(roots))
	for _, id := range roots {
		if node, ok := build(id); ok {
			out = append(out, node)
		}
	}
	return out
}
