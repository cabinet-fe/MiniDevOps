package service_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	"bedrock/internal/resource/model"
	"bedrock/internal/resource/repository"
	"bedrock/internal/resource/service"
)

func setupPAT(t *testing.T) *service.PATService {
	t.Helper()
	gdb, err := db.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "pat.sqlite"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.Up(context.Background(), gdb, migration.Driver("sqlite")); err != nil {
		t.Fatalf("migration: %v", err)
	}
	return service.NewPATService(repository.NewPATRepository(gdb))
}

func TestPATPlaintextOnceAndScopes(t *testing.T) {
	pats := setupPAT(t)
	created, err := pats.Create(1, service.CreatePATInput{
		Name: "t", Scopes: []string{model.ScopeSkillsRead},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(created.Token, "br_pat_") {
		t.Fatalf("unexpected token %s", created.Token)
	}
	list, _, err := pats.List(1, 1, 20)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
	// Metadata list must not include plaintext.
	if strings.Contains(list[0].TokenHash, created.Token) {
		t.Fatal("plaintext must not be stored")
	}
	if _, _, err := pats.ValidateBearer("br_pat_deadbeef"); err == nil {
		t.Fatal("invalid PAT must fail")
	}
	uid, scopes, err := pats.ValidateBearer(created.Token)
	if err != nil || uid != 1 {
		t.Fatalf("valid PAT: %v uid=%d", err, uid)
	}
	if err := pats.RequireScope(scopes, model.ScopeAgentsRun); err == nil {
		t.Fatal("wrong scope must 403")
	}
	if err := pats.RequireScope(scopes, model.ScopeSkillsRead); err != nil {
		t.Fatal(err)
	}
	if err := pats.Delete(1, created.Metadata.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := pats.ValidateBearer(created.Token); err == nil {
		t.Fatal("deleted PAT must be invalid")
	}
	list, _, err = pats.List(1, 1, 20)
	if err != nil || len(list) != 0 {
		t.Fatalf("deleted PAT must be removed from list: %v %#v", err, list)
	}
}

func TestPATUserScopedDelete(t *testing.T) {
	pats := setupPAT(t)
	created, err := pats.Create(1, service.CreatePATInput{
		Name: "mine", Scopes: []string{model.ScopeAgentsRun},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := pats.Delete(2, created.Metadata.ID); err == nil {
		t.Fatal("other user must not delete someone else's PAT")
	}
	if err := pats.Delete(1, created.Metadata.ID); err != nil {
		t.Fatal(err)
	}
}
