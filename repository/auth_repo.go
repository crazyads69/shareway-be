package repository

import (
	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IAuthRepository defines the interface for authentication-related database operations
type IAuthRepository interface {
	UserExistsByPhone(phoneNumber string) (bool, error)
	CreateUserByPhone(phoneNumber, fullName string) (uuid.UUID, string, error)
	GetUserIDByPhone(phoneNumber string) (uuid.UUID, error)
	ActivateUser(phoneNumber string) error
	GetUserByPhone(phoneNumber string) (migration.User, error)
	SaveCCCDInfo(cccdEncrypted string, userID uuid.UUID) error
	VerifyUser(phoneNumber string) error
	SaveSession(phoneNumber string, accessToken string, refreshToken string, userID uuid.UUID) error
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
	var count int64
	err := r.db.Model(&migration.User{}).
		Where("phone_number = ?", phoneNumber).
		Count(&count).
		Error

	return count > 0, err
}

// CreateUserByPhone creates a new user with the given phone number and full name
func (r *AuthRepository) CreateUserByPhone(phoneNumber, fullName string) (uuid.UUID, string, error) {
	user := migration.User{
		PhoneNumber: phoneNumber,
		FullName:    fullName,
	}
	err := r.db.Create(&user).Error
	if err != nil {
		return uuid.Nil, "", err
	}
	return user.ID, user.FullName, nil
}

// GetUserIDByPhone retrieves the user ID associated with the given phone number
func (r *AuthRepository) GetUserIDByPhone(phoneNumber string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.Model(&migration.User{}).
		Select("id").
		Where("phone_number = ?", phoneNumber).
		First(&userID).
		Error

	return userID, err
}

// ActivateUser updates the user status to activated
func (r *AuthRepository) ActivateUser(phoneNumber string) error {
	return r.db.Model(&migration.User{}).
		Where("phone_number = ?", phoneNumber).
		Update("is_activated", true).
		Error
}

// GetUserByPhone retrieves the user associated with the given phone number
func (r *AuthRepository) GetUserByPhone(phoneNumber string) (migration.User, error) {
	var user migration.User
	err := r.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	return user, err
}

// SaveCCCDInfo saves the encrypted CCCD information to the database
func (r *AuthRepository) SaveCCCDInfo(cccdEncrypted string, userID uuid.UUID) error {
	// Update the user's CCCD information (CCCDNumber in the User model)
	return r.db.Model(&migration.User{}).Where("id = ?", userID).Update("cccd_number", cccdEncrypted).Error
}

// VerifyUser updates the user status to verified
func (r *AuthRepository) VerifyUser(phoneNumber string) error {
	return r.db.Model(&migration.User{}).
		Where("phone_number = ?", phoneNumber).
		Update("is_verified", true).
		Error
}

// SaveSession saves the access token and refresh token to the database
func (r *AuthRepository) SaveSession(phoneNumber string, accessToken string, refreshToken string, userID uuid.UUID) error {
	// Insert to the PasetoToken table
	token := migration.PasetoToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
	}
	return r.db.Create(&token).Error
}

// Ensure AuthRepository implements IAuthRepository
var _ IAuthRepository = (*AuthRepository)(nil)
