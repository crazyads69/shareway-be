package repository

import (
	"errors"
	"time"

	"shareway/db/migration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTPRepository handles OTP-related database operations
type OTPRepository struct {
	db *gorm.DB
}

// NewOTPRepository creates a new OTPRepository instance
func NewOTPRepository(db *gorm.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

// IOTPRepository defines the interface for OTP operations
type IOTPRepository interface {
	StoreOTP(phoneNumber string, otp string, userID uuid.UUID) error
	DeleteOTP(userID uuid.UUID) error
	CleanupExpiredOTPs() error
}

// StoreOTP stores or updates an OTP for a given user
func (repo *OTPRepository) StoreOTP(phoneNumber string, otp string, userID uuid.UUID) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// Clean up expired OTPs before proceeding
		if err := repo.CleanupExpiredOTPs(); err != nil {
			return err
		}

		var existingOTP migration.OTP
		// Check for an existing active OTP for the user
		err := tx.Where("user_id = ? AND expires_at > ?", userID, time.Now()).First(&existingOTP).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// If no active OTP found, create a new one
				newOTP := migration.OTP{
					ID:          uuid.New(),
					PhoneNumber: phoneNumber,
					Code:        otp,
					ExpiresAt:   time.Now().Add(10 * time.Minute),
					UserID:      userID,
					Retry:       0,
				}
				return tx.Create(&newOTP).Error
			}
			return err
		}

		// Check if the maximum retry limit has been reached
		if existingOTP.Retry >= 2 {
			return errors.New("maximum OTP requests reached")
		}

		// Update existing OTP
		existingOTP.Code = otp
		existingOTP.ExpiresAt = time.Now().Add(10 * time.Minute)
		existingOTP.Retry++
		return tx.Save(&existingOTP).Error
	})
}

// DeleteOTP removes all OTPs associated with a given user ID
func (repo *OTPRepository) DeleteOTP(userID uuid.UUID) error {
	return repo.db.Where("user_id = ?", userID).Delete(&migration.OTP{}).Error
}

// CleanupExpiredOTPs removes all expired OTPs from the database
func (repo *OTPRepository) CleanupExpiredOTPs() error {
	return repo.db.Where("expires_at <= ?", time.Now()).Delete(&migration.OTP{}).Error
}

// Ensure OTPRepository implements IOTPRepository interface
var _ IOTPRepository = (*OTPRepository)(nil)
