package service

import (
	"errors"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(username, password, displayName, role, email string) (*model.User, error) {
	hash, err := pkg.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		DisplayName:  displayName,
		Role:         role,
		Email:        email,
		IsActive:     true,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) Authenticate(username, password string) (*model.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if !user.IsActive {
		return nil, errors.New("账户已被禁用")
	}
	if !pkg.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}
	return user, nil
}

func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.List(page, pageSize)
}

func (s *UserService) Update(user *model.User) error {
	return s.repo.Update(user)
}

func (s *UserService) Delete(id uint) error {
	// Check if this is the last admin
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		count, _ := s.repo.CountByRole("admin")
		if count <= 1 {
			return errors.New("不能删除最后一个管理员")
		}
	}
	return s.repo.Delete(id)
}

func (s *UserService) UpdateProfile(userID uint, displayName, email, avatar string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if displayName != "" {
		user.DisplayName = displayName
	}
	if email != "" {
		user.Email = email
	}
	if avatar != "" {
		user.Avatar = avatar
	}
	return s.repo.Update(user)
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if !pkg.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("原密码错误")
	}
	hash, err := pkg.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.repo.Update(user)
}
