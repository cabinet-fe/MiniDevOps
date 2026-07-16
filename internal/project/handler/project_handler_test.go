package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	authmodel "bedrock/internal/auth/model"
	authrepo "bedrock/internal/auth/repository"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	"bedrock/internal/platform/seed"
	projectrepo "bedrock/internal/project/repository"
	projectservice "bedrock/internal/project/service"
	rbacrepo "bedrock/internal/rbac/repository"
	rbacservice "bedrock/internal/rbac/service"
	storagerepo "bedrock/internal/storage/repository"
	storageservice "bedrock/internal/storage/service"
)

func TestPublishDocNodeRequiresExpectedVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, service := newProjectHandlerForTest(t)
	owner := projectservice.NewAccessContext(1, true, nil)
	project, err := service.CreateProject(owner, projectservice.CreateProjectInput{Name: "Docs", Slug: "docs"})
	if err != nil {
		t.Fatal(err)
	}
	draft := "draft"
	node, err := service.CreateDocNode(owner, project.ID, projectservice.DocNodeInput{
		Kind: "doc", Name: "guide.md", DraftContent: &draft,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := publishDocNodeRequest(t, handler, project.ID, node.ID, `{}`); got != http.StatusBadRequest {
		t.Fatalf("missing expected_version status = %d, want %d", got, http.StatusBadRequest)
	}
	if got := publishDocNodeRequest(t, handler, project.ID, node.ID, `{"expected_version":1}`); got != http.StatusConflict {
		t.Fatalf("stale expected_version status = %d, want %d", got, http.StatusConflict)
	}
}

func TestListRequirementStatusesAllowsLeastPrivilegeMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, service, gdb := newProjectHandlerForTestWithDB(t)
	if err := seed.EnsureRBACResources(gdb); err != nil {
		t.Fatal(err)
	}

	users := authrepo.NewUserRepository(gdb)
	roles := rbacrepo.NewRoleRepository(gdb)
	roleService := rbacservice.NewRoleService(roles)
	member := &authmodel.User{Username: "requirement_reader", PasswordHash: "hash", IsActive: true}
	if err := users.Create(member); err != nil {
		t.Fatal(err)
	}
	role, err := roleService.Create("需求只读", "requirement_reader", "", []string{"project.requirements:view"})
	if err != nil {
		t.Fatal(err)
	}
	if err := roleService.SetUserRoles(member.ID, []uint{role.ID}); err != nil {
		t.Fatal(err)
	}

	owner := projectservice.NewAccessContext(999, true, nil)
	project, err := service.CreateProject(owner, projectservice.CreateProjectInput{Name: "Requirements", Slug: "requirements"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddMember(owner, project.ID, projectservice.MemberInput{
		UserID: member.ID,
		Role:   "readonly",
	}); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/projects/meta/requirement-statuses", nil)
	c.Set("user_id", member.ID)
	c.Set("is_super_admin", false)
	handler.ListRequirementStatuses(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"backlog"`)) {
		t.Fatalf("response must include baseline requirement statuses: %s", recorder.Body.String())
	}
}

func TestGenerateDocsRemainsNotImplemented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, service := newProjectHandlerForTest(t)
	owner := projectservice.NewAccessContext(1, true, nil)
	project, err := service.CreateProject(owner, projectservice.CreateProjectInput{Name: "Docs", Slug: "generate-docs"})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/1/docs/generate", nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(project.ID), 10)}}
	c.Set("user_id", uint(1))
	c.Set("is_super_admin", true)
	handler.GenerateDocs(c)

	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func publishDocNodeRequest(t *testing.T, handler *ProjectHandler, projectID, nodeID uint, body string) int {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/projects/1/docs/1/publish",
		bytes.NewBufferString(body),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "nodeID", Value: "1"},
	}
	c.Params[0].Value = strconv.FormatUint(uint64(projectID), 10)
	c.Params[1].Value = strconv.FormatUint(uint64(nodeID), 10)
	c.Set("user_id", uint(1))
	c.Set("is_super_admin", true)
	handler.PublishDocNode(c)
	return recorder.Code
}

func newProjectHandlerForTest(t *testing.T) (*ProjectHandler, *projectservice.ProjectService) {
	handler, service, _ := newProjectHandlerForTestWithDB(t)
	return handler, service
}

func newProjectHandlerForTestWithDB(t *testing.T) (*ProjectHandler, *projectservice.ProjectService, *gorm.DB) {
	t.Helper()
	gdb, err := db.Open(&config.DatabaseConfig{Driver: "sqlite", Path: t.TempDir() + "/project-handler.sqlite"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
	if err := migration.Up(context.Background(), gdb, migration.Driver("sqlite")); err != nil {
		t.Fatal(err)
	}
	storage, err := storageservice.NewStorageService(
		storagerepo.NewStorageRepository(gdb),
		t.TempDir(),
		storageservice.Limits{},
	)
	if err != nil {
		t.Fatal(err)
	}
	service := projectservice.NewProjectService(projectrepo.NewProjectRepository(gdb), storage)
	permissions := rbacservice.NewPermissionService(
		rbacrepo.NewRoleRepository(gdb),
		rbacrepo.NewResourceRepository(gdb),
	)
	return NewProjectHandler(service, permissions), service, gdb
}
