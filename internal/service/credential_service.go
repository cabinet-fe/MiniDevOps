package service

import (
	"errors"
	"strings"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
	"gorm.io/gorm"
)

type CredentialOption struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type CredentialService struct {
	repo        *repository.CredentialRepository
	projectRepo *repository.ProjectRepository
	userRepo    *repository.UserRepository
}

func NewCredentialService(
	repo *repository.CredentialRepository,
	projectRepo *repository.ProjectRepository,
	userRepo *repository.UserRepository,
) *CredentialService {
	return &CredentialService{
		repo:        repo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (s *CredentialService) Create(credential *model.Credential) error {
	credential.Name = strings.TrimSpace(credential.Name)
	credential.Type = normalizeCredentialType(credential.Type)
	credential.Username = strings.TrimSpace(credential.Username)
	credential.Description = strings.TrimSpace(credential.Description)
	if err := validateCredentialForCreate(credential); err != nil {
		return err
	}
	encrypted, err := pkg.Encrypt(credential.Password)
	if err != nil {
		return err
	}
	credential.Password = encrypted
	return s.repo.Create(credential)
}

func (s *CredentialService) Update(credential *model.Credential) error {
	existing, err := s.repo.FindByID(credential.ID)
	if err != nil {
		return err
	}

	existing.Name = strings.TrimSpace(credential.Name)
	existing.Type = normalizeCredentialType(credential.Type)
	existing.Username = strings.TrimSpace(credential.Username)
	existing.Description = strings.TrimSpace(credential.Description)
	if err := validateCredentialForUpdate(existing); err != nil {
		return err
	}

	if strings.TrimSpace(credential.Password) == "" {
		// keep old secret
	} else {
		encrypted, err := pkg.Encrypt(credential.Password)
		if err != nil {
			return err
		}
		existing.Password = encrypted
	}
	return s.repo.Update(existing)
}

func (s *CredentialService) Delete(id uint) error {
	count, err := s.projectRepo.CountByCredentialID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该凭证正在被项目引用，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *CredentialService) GetByID(id uint) (*model.Credential, error) {
	credential, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	masked := maskCredential(*credential)
	s.fillCreatorNames([]model.Credential{masked}, &masked)
	return &masked, nil
}

func (s *CredentialService) GetByIDWithSecret(id uint) (*model.Credential, error) {
	credential, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if credential.Password != "" {
		plain, err := pkg.Decrypt(credential.Password)
		if err != nil {
			return nil, err
		}
		credential.Password = plain
	}
	return credential, nil
}

func (s *CredentialService) ListByUser(userID uint, role string) ([]model.Credential, error) {
	var (
		credentials []model.Credential
		err         error
	)
	if role == "admin" {
		credentials, err = s.repo.FindAll()
	} else {
		credentials, err = s.repo.FindByCreator(userID)
	}
	if err != nil {
		return nil, err
	}

	items := make([]model.Credential, 0, len(credentials))
	for _, credential := range credentials {
		items = append(items, maskCredential(credential))
	}
	s.fillCreatorNames(items, nil)
	return items, nil
}

func (s *CredentialService) ListForSelect(userID uint, role string) ([]CredentialOption, error) {
	items, err := s.ListByUser(userID, role)
	if err != nil {
		return nil, err
	}
	result := make([]CredentialOption, 0, len(items))
	for _, item := range items {
		result = append(result, CredentialOption{
			ID:   item.ID,
			Name: item.Name,
			Type: item.Type,
		})
	}
	return result, nil
}

func (s *CredentialService) CanUseCredential(credentialID, userID uint, role string) (bool, error) {
	credential, err := s.repo.FindByID(credentialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if role == "admin" {
		return true, nil
	}
	return credential.CreatedBy == userID, nil
}

func (s *CredentialService) fillCreatorNames(items []model.Credential, single *model.Credential) {
	names := make(map[uint]string, len(items))
	for _, item := range items {
		if item.CreatedBy == 0 {
			continue
		}
		if _, exists := names[item.CreatedBy]; exists {
			continue
		}
		user, err := s.userRepo.FindByID(item.CreatedBy)
		if err != nil {
			names[item.CreatedBy] = ""
			continue
		}
		display := strings.TrimSpace(user.DisplayName)
		if display == "" {
			display = user.Username
		}
		names[item.CreatedBy] = display
	}
	if single != nil {
		single.CreatorName = names[single.CreatedBy]
		return
	}
	for i := range items {
		items[i].CreatorName = names[items[i].CreatedBy]
	}
}

func maskCredential(credential model.Credential) model.Credential {
	credential.HasSecret = strings.TrimSpace(credential.Password) != ""
	credential.Password = ""
	return credential
}

func normalizeCredentialType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "token":
		return "token"
	default:
		return "password"
	}
}

func validateCredentialForCreate(credential *model.Credential) error {
	if credential.Name == "" {
		return errors.New("凭证名称不能为空")
	}
	if credential.Password == "" {
		return errors.New("密码或 Token 不能为空")
	}
	if credential.Type == "password" && credential.Username == "" {
		return errors.New("用户名不能为空")
	}
	if credential.Type != "password" && credential.Type != "token" {
		return errors.New("不支持的凭证类型")
	}
	return nil
}

func validateCredentialForUpdate(credential *model.Credential) error {
	if credential.Name == "" {
		return errors.New("凭证名称不能为空")
	}
	if credential.Type == "password" && credential.Username == "" {
		return errors.New("用户名不能为空")
	}
	if credential.Type != "password" && credential.Type != "token" {
		return errors.New("不支持的凭证类型")
	}
	return nil
}
