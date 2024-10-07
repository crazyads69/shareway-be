package repository

import (
	"errors"
	"fmt"
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
	UserExistsByEmail(email string) (bool, error)
	CreateUser(phoneNumber string, fullName string, email string) (uuid.UUID, error)
	GetUserByEmail(email string) (migration.User, error)
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

// UserExistsByEmail checks if a user exists with the given email
// UserExistsByEmail checks if a user exists with the given email
func (r *AuthRepository) UserExistsByEmail(email string) (bool, error) {
	// Use a more efficient query that doesn't need to count all matching rows
	var exists bool
	err := r.db.Model(&migration.User{}).
		Select("1").
		Where("email = ?", email).
		Limit(1).
		Find(&exists).
		Error

	// Return true if a user was found, false otherwise
	// If there's an error, 'exists' will be false and the error will be returned
	return exists, err
}

// CreateUser creates a new user with the given phone number, full name, and email
// CreateUserByPhoneAndEmail creates a new user with the given phone number, full name, and email
// It returns the newly created user's UUID and any error that occurred during the process
func (r *AuthRepository) CreateUser(phoneNumber, fullName, email string) (uuid.UUID, error) {
	// Create a new User instance with the provided information
	user := migration.User{
		PhoneNumber: phoneNumber,
		FullName:    fullName,
		Email:       email,
	}

	// Attempt to insert the new user into the database
	result := r.db.Create(&user)
	if result.Error != nil {
		// If there's an error during creation, return a nil UUID and the error
		return uuid.Nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	// If successful, return the new user's ID
	return user.ID, nil
}

// GetUserByEmail retrieves the user associated with the given email
// GetUserByEmail retrieves a user by their email address
// It returns the user and any error encountered during the process
func (r *AuthRepository) GetUserByEmail(email string) (migration.User, error) {
	var user migration.User

	// Query the database for a user with the given email
	// Use First() to fetch only one record and stop searching after finding it
	err := r.db.Where("email = ?", email).First(&user).Error

	// Handle potential errors
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no user is found, return an empty user and a custom error
			return migration.User{}, fmt.Errorf("user with email %s not found", email)
		}
		// For any other error, return an empty user and the error
		return migration.User{}, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Return the found user and nil error
	return user, nil
}

// Ensure AuthRepository implements IAuthRepository
var _ IAuthRepository = (*AuthRepository)(nil)
