package repository

import (
	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IAuthRepository defines the interface for authentication-related database operations
type IAuthRepository interface {
	UserExistsByPhone(phoneNumber string) (bool, error)
	CreateUserByPhone(phoneNumber string) (uuid.UUID, error)
}

// AuthRepository implements IAuthRepository
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new instance of AuthRepository
func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{db: db}
}

// UserExistsByPhone checks if a user exists with the given phone number
func (r *AuthRepository) UserExistsByPhone(phoneNumber string) (bool, error) {
	var exists bool
	err := r.db.Model(&migration.User{}).
		Select("1").
		Where("phone_number = ?", phoneNumber).
		Limit(1).
		Find(&exists).
		Error

	if err != nil {
		return false, err
	}
	return exists, nil
}

// CreateUserByPhone creates a new user with the given phone number
func (r *AuthRepository) CreateUserByPhone(phoneNumber string) (uuid.UUID, error) {
	user := migration.User{
		ID:          uuid.New(),
		PhoneNumber: phoneNumber,
	}
	err := r.db.Create(&user).Error
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

// Ensure AuthRepository implements IAuthRepository
var _ IAuthRepository = (*AuthRepository)(nil)
