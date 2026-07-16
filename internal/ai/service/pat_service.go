package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"bedrock/internal/ai/model"
	"bedrock/internal/ai/repository"
)

var (
	ErrPATInvalid     = errors.New("invalid or expired token")
	ErrPATWrongScope  = errors.New("token scope insufficient")
	ErrPATBadScope    = errors.New("scope 仅允许 skills:read 与 agents:run")
)

type PATService struct {
	repo  *repository.AIRepository
	audit AuditWriter
}

func NewPATService(repo *repository.AIRepository, audit ...AuditWriter) *PATService {
	svc := &PATService{repo: repo}
	if len(audit) > 0 {
		svc.audit = audit[0]
	}
	return svc
}

type CreatePATInput struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type CreatePATResult struct {
	Token   string                    `json:"token"` // plaintext, only in create response
	Metadata model.PersonalAccessToken `json:"metadata"`
}

func (s *PATService) Create(userID uint, in CreatePATInput) (*CreatePATResult, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errors.New("名称不能为空")
	}
	scopes, err := normalizeScopes(in.Scopes)
	if err != nil {
		return nil, err
	}
	plain, err := generatePATPlaintext()
	if err != nil {
		return nil, err
	}
	hash := hashToken(plain)
	scopesJSON, _ := json.Marshal(scopes)
	item := &model.PersonalAccessToken{
		UserID:      userID,
		Name:        name,
		TokenPrefix: plain[:12],
		TokenHash:   hash,
		ScopesJSON:  string(scopesJSON),
		Scopes:      scopes,
		ExpiresAt:   in.ExpiresAt,
	}
	if err := s.repo.CreatePAT(item); err != nil {
		return nil, err
	}
	if s.audit != nil {
		_ = s.audit.Write(userID, "", "pat_create", "personal_access_token", fmt.Sprintf("%d", item.ID),
			fmt.Sprintf("name=%s scopes=%v", name, scopes), "")
	}
	return &CreatePATResult{Token: plain, Metadata: *item}, nil
}

func (s *PATService) List(userID uint) ([]model.PersonalAccessToken, error) {
	items, err := s.repo.ListPATs(userID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		decodeScopes(&items[i])
	}
	return items, nil
}

func (s *PATService) Delete(userID uint, id uint) error {
	token, err := s.repo.FindPAT(id)
	if err != nil {
		return err
	}
	if token.UserID != userID {
		return ErrPATInvalid
	}
	if err := s.repo.DeletePAT(id); err != nil {
		return err
	}
	if s.audit != nil {
		_ = s.audit.Write(userID, "", "pat_delete", "personal_access_token", fmt.Sprintf("%d", id),
			fmt.Sprintf("name=%s", token.Name), "")
	}
	return nil
}

// ValidateBearer returns userID and scopes for a valid PAT; otherwise ErrPATInvalid.
func (s *PATService) ValidateBearer(raw string) (userID uint, scopes []string, err error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "br_pat_") {
		return 0, nil, ErrPATInvalid
	}
	token, err := s.repo.FindPATByHash(hashToken(raw))
	if err != nil {
		return 0, nil, ErrPATInvalid
	}
	if token.RevokedAt != nil {
		return 0, nil, ErrPATInvalid
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now().UTC()) {
		return 0, nil, ErrPATInvalid
	}
	decodeScopes(token)
	now := time.Now().UTC()
	token.LastUsedAt = &now
	_ = s.repo.UpdatePAT(token)
	return token.UserID, token.Scopes, nil
}

func (s *PATService) RequireScope(scopes []string, required string) error {
	for _, sc := range scopes {
		if sc == required {
			return nil
		}
	}
	return ErrPATWrongScope
}

func normalizeScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, ErrPATBadScope
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(scopes))
	for _, sc := range scopes {
		sc = strings.TrimSpace(sc)
		if sc != model.ScopeSkillsRead && sc != model.ScopeAgentsRun {
			return nil, ErrPATBadScope
		}
		if !seen[sc] {
			seen[sc] = true
			out = append(out, sc)
		}
	}
	return out, nil
}

func decodeScopes(token *model.PersonalAccessToken) {
	_ = json.Unmarshal([]byte(token.ScopesJSON), &token.Scopes)
}

func generatePATPlaintext() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "br_pat_" + hex.EncodeToString(buf), nil
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
