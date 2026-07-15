package service_test

import (
	"context"
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	authmodel "bedrock/internal/auth/model"
	authrepo "bedrock/internal/auth/repository"
	"bedrock/internal/pkg"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	"bedrock/internal/platform/seed"
	"bedrock/internal/rbac"
	"bedrock/internal/rbac/model"
	rbacrepo "bedrock/internal/rbac/repository"
	"bedrock/internal/rbac/service"
)

func setupRBAC(t *testing.T) (*service.PermissionService, *service.RoleService, *service.ResourceService, *authrepo.UserRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	gdb, err := db.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "rbac.sqlite"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
	if err := migration.Up(context.Background(), gdb, "sqlite"); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureRBACResources(gdb); err != nil {
		t.Fatal(err)
	}

	roles := rbacrepo.NewRoleRepository(gdb)
	resources := rbacrepo.NewResourceRepository(gdb)
	users := authrepo.NewUserRepository(gdb)
	return service.NewPermissionService(roles, resources),
		service.NewRoleService(roles),
		service.NewResourceService(resources),
		users
}

func TestPermissionUnion(t *testing.T) {
	perm, roles, _, users := setupRBAC(t)

	hash, _ := pkg.HashPassword("pass")
	u := &authmodel.User{Username: "u1", PasswordHash: hash, IsActive: true}
	if err := users.Create(u); err != nil {
		t.Fatal(err)
	}

	r1, err := roles.Create("A", "role_a", "", []string{"system.users:view", "cicd.repositories:view"})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := roles.Create("B", "role_b", "", []string{"system.roles:view", "cicd.repositories:create"})
	if err != nil {
		t.Fatal(err)
	}
	if err := roles.SetUserRoles(u.ID, []uint{r1.ID, r2.ID}); err != nil {
		t.Fatal(err)
	}

	codes, err := perm.ResolvePermissions(u.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	set := rbac.ToSet(codes)
	for _, want := range []string{
		"system.users:view", "cicd.repositories:view", "system.roles:view", "cicd.repositories:create",
	} {
		if !rbac.HasPermission(set, want) {
			t.Fatalf("missing %s in %v", want, codes)
		}
	}
}

func TestOpsHardGate(t *testing.T) {
	perm, roles, _, users := setupRBAC(t)

	hash, _ := pkg.HashPassword("pass")
	u := &authmodel.User{Username: "opsfan", PasswordHash: hash, IsActive: true}
	if err := users.Create(u); err != nil {
		t.Fatal(err)
	}
	r, err := roles.Create("OpsMistaken", "ops_mistaken", "", []string{
		"ops:view", "ops.processes:view", "ops.processes:delete", "system.users:view",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := roles.SetUserRoles(u.ID, []uint{r.ID}); err != nil {
		t.Fatal(err)
	}

	codes, err := perm.ResolvePermissions(u.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range codes {
		if rbac.IsOpsPermission(c) {
			t.Fatalf("ops permission leaked to non-super: %s", c)
		}
	}
	if err := perm.CheckAccess(u.ID, false, "ops.processes:view"); err == nil || !service.IsForbidden(err) {
		t.Fatalf("expected ops hard gate 403, got %v", err)
	}
	if err := perm.CheckAccess(u.ID, false, "system.users:view"); err != nil {
		t.Fatalf("system.users:view should pass: %v", err)
	}
	if err := perm.CheckAccess(1, true, "ops.processes:view"); err != nil {
		t.Fatalf("super-admin should pass ops: %v", err)
	}
}

func TestMenuTrimAndParentFill(t *testing.T) {
	perm, roles, _, users := setupRBAC(t)

	hash, _ := pkg.HashPassword("pass")
	u := &authmodel.User{Username: "viewer", PasswordHash: hash, IsActive: true}
	if err := users.Create(u); err != nil {
		t.Fatal(err)
	}
	// Only leaf view — parent system should auto-fill.
	r, err := roles.Create("Viewer", "viewer", "", []string{"system.users:view"})
	if err != nil {
		t.Fatal(err)
	}
	if err := roles.SetUserRoles(u.ID, []uint{r.ID}); err != nil {
		t.Fatal(err)
	}

	menus, err := perm.TrimMenus(u.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(menus) != 1 || menus[0].Path != "system" {
		t.Fatalf("expected system parent, got %+v", menus)
	}
	if len(menus[0].Children) != 1 || menus[0].Children[0].Path != "system.users" {
		t.Fatalf("expected system.users leaf, got %+v", menus[0].Children)
	}

	// No :view → empty menus
	r2, err := roles.Create("NoView", "noview", "", []string{"system.users:create"})
	if err != nil {
		t.Fatal(err)
	}
	u2 := &authmodel.User{Username: "noview", PasswordHash: hash, IsActive: true}
	if err := users.Create(u2); err != nil {
		t.Fatal(err)
	}
	_ = roles.SetUserRoles(u2.ID, []uint{r2.ID})
	menus2, err := perm.TrimMenus(u2.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(menus2) != 0 {
		t.Fatalf("expected empty menus without :view, got %+v", menus2)
	}
	if err := perm.CheckAccess(u2.ID, false, "system.users:view"); err == nil {
		t.Fatal("expected 403 without :view")
	}
}

func TestMenuIconRejectOver32KB(t *testing.T) {
	_, _, resources, _ := setupRBAC(t)

	// Find top-level system menu id
	tree, err := resources.ListMenusTree()
	if err != nil {
		t.Fatal(err)
	}
	var systemID uint
	for _, n := range tree {
		if n.Path == "system" {
			systemID = n.ID
			break
		}
	}
	if systemID == 0 {
		t.Fatal("system menu not found")
	}

	big := make([]byte, rbac.MaxMenuIconBytes+1)
	for i := range big {
		big[i] = 'A'
	}
	payload := base64.StdEncoding.EncodeToString(big)
	_, err = resources.UpdateMenuIcon(systemID, payload, "image/png")
	if err == nil || !strings.Contains(err.Error(), "32KB") {
		t.Fatalf("expected 32KB reject, got %v", err)
	}

	ok := make([]byte, 16)
	_, err = resources.UpdateMenuIcon(systemID, base64.StdEncoding.EncodeToString(ok), "image/png")
	if err != nil {
		t.Fatalf("small icon should pass: %v", err)
	}
}

func TestIsOpsPath(t *testing.T) {
	if !rbac.IsOpsPath("ops") || !rbac.IsOpsPath("ops.processes") {
		t.Fatal("ops paths")
	}
	if rbac.IsOpsPath("system") || rbac.IsOpsPath("options") {
		t.Fatal("false positive ops")
	}
	_ = model.ResourceTypeMenu
}
