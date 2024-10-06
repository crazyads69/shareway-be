package repository

import (
	"errors"
	"time"

	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

type IAuthRepository interface {
	StoreOTP(phoneNumber string, otp string, userID uuid.UUID) error
	GetLatestOTP(phoneNumber string) (*migration.OTP, error)
	DeleteOTP(id uuid.UUID) error
	UpdateRetry(id uuid.UUID) error
	CleanupExpiredOTPs() error
}

func (repo *AuthRepository) StoreOTP(phoneNumber string, otp string, userID uuid.UUID) error {
	// First, clean up any expired OTPs
	if err := repo.CleanupExpiredOTPs(); err != nil {
		return err
	}

	// Check for an existing active OTP
	existingOTP, err := repo.GetLatestOTP(phoneNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existingOTP != nil {
		// Update the existing OTP
		existingOTP.Code = otp
		existingOTP.ExpiresAt = time.Now().Add(10 * time.Minute) // Extend expiry by 10 minutes
		return repo.db.Save(existingOTP).Error
	}

	// If no active OTP exists, create a new one
	newOTP := migration.OTP{
		ID:          uuid.New(),
		PhoneNumber: phoneNumber,
		Code:        otp,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		UserID:      userID,
	}
	return repo.db.Create(&newOTP).Error
}

func (repo *AuthRepository) GetLatestOTP(phoneNumber string) (*migration.OTP, error) {
	var otp migration.OTP
	err := repo.db.Where("phone_number = ? AND expires_at > ?", phoneNumber, time.Now()).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (repo *AuthRepository) DeleteOTP(id uuid.UUID) error {
	return repo.db.Delete(&migration.OTP{}, id).Error
}

func (repo *AuthRepository) UpdateRetry(id uuid.UUID) error {
	return repo.db.Model(&migration.OTP{}).
		Where("id = ?", id).
		UpdateColumn("retry", gorm.Expr("retry + ?", 1)).
		Error
}

func (repo *AuthRepository) CleanupExpiredOTPs() error {
	return repo.db.Where("expires_at <= ?", time.Now()).Delete(&migration.OTP{}).Error
}
