package service

import (
	"shareway/repository"

	"github.com/google/uuid"
)

type UsersService struct {
	repo repository.IAuthRepository
}

type IUsersService interface {
	UserExistsByPhone(phoneNumber string) (bool, error)
	CreateUserByPhone(phoneNumber string) (uuid.UUID, error)
}

func NewUsersService(repo repository.IAuthRepository) *UsersService {
	return &UsersService{
		repo: repo,
	}
}

func (s *UsersService) UserExistsByPhone(phoneNumber string) (bool, error) {
	return s.repo.UserExistsByPhone(phoneNumber)
}

func (s *UsersService) CreateUserByPhone(phoneNumber string) (uuid.UUID, error) {
	return s.repo.CreateUserByPhone(phoneNumber)
}
