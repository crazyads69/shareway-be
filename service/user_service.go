package service

import (
	"shareway/infra/db/migration"
	"shareway/repository"

	"github.com/google/uuid"
)

// IUsersService defines the interface for user-related business logic operations
type IUsersService interface {
	UserExistsByPhone(phoneNumber string) (bool, error)
	CreateUserByPhone(phoneNumber, fullName string) (uuid.UUID, string, error)
	GetUserIDByPhone(phoneNumber string) (uuid.UUID, error)
	ActivateUser(phoneNumber string) error
	GetUserByPhone(phoneNumber string) (migration.User, error)
}

// UsersService implements IUsersService and handles user-related business logic
type UsersService struct {
	repo repository.IAuthRepository
}

// NewUsersService creates a new instance of UsersService
func NewUsersService(repo repository.IAuthRepository) IUsersService {
	return &UsersService{
		repo: repo,
	}
}

// UserExistsByPhone checks if a user exists with the given phone number
func (s *UsersService) UserExistsByPhone(phoneNumber string) (bool, error) {
	return s.repo.UserExistsByPhone(phoneNumber)
}

// CreateUserByPhone creates a new user with the given phone number and full name
func (s *UsersService) CreateUserByPhone(phoneNumber, fullName string) (uuid.UUID, string, error) {
	return s.repo.CreateUserByPhone(phoneNumber, fullName)
}

// GetUserIDByPhone retrieves the user ID associated with the given phone number
func (s *UsersService) GetUserIDByPhone(phoneNumber string) (uuid.UUID, error) {
	return s.repo.GetUserIDByPhone(phoneNumber)
}

// ActivateUser activates the user account associated with the given phone number
func (s *UsersService) ActivateUser(phoneNumber string) error {
	return s.repo.ActivateUser(phoneNumber)
}

// GetUserByPhone retrieves the user associated with the given phone number
func (s *UsersService) GetUserByPhone(phoneNumber string) (migration.User, error) {
	return s.repo.GetUserByPhone(phoneNumber)
}

// Ensure UsersService implements IUsersService
var _ IUsersService = (*UsersService)(nil)
