package services

import (
	"database/sql"
	"errors"

	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/repositories"
)

var ErrRoleNotFound = errors.New("role not found")

type UserService interface {
	EnsureUserWithDefaultRole(auth0UserID, defaultRole string) error
	ListUsersWithRoles() ([]models.UserWithRole, error)
	SetUserRole(auth0UserID, roleName string) error
	DeleteUser(auth0UserID string) error
	IsUserInRole(auth0UserID, roleName string) (bool, error)
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) EnsureUserWithDefaultRole(auth0UserID, defaultRole string) error {
	if err := s.repo.EnsureUser(auth0UserID); err != nil {
		return err
	}
	role, err := s.repo.GetUserRole(auth0UserID)
	if err != nil {
		return err
	}
	if role != "" {
		return nil
	}
	return s.SetUserRole(auth0UserID, defaultRole)
}

func (s *userService) ListUsersWithRoles() ([]models.UserWithRole, error) {
	return s.repo.ListUsersWithRoles()
}

func (s *userService) SetUserRole(auth0UserID, roleName string) error {
	if err := s.repo.EnsureUser(auth0UserID); err != nil {
		return err
	}
	if err := s.repo.SetUserRole(auth0UserID, roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRoleNotFound
		}
		return err
	}
	return nil
}

func (s *userService) DeleteUser(auth0UserID string) error {
	return s.repo.DeleteUser(auth0UserID)
}

func (s *userService) IsUserInRole(auth0UserID, roleName string) (bool, error) {
	return s.repo.IsUserInRole(auth0UserID, roleName)
}
