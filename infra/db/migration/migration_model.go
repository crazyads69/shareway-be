package migration

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	PhoneNumber string         `gorm:"uniqueIndex;not null"`
	Email       string         `gorm:"uniqueIndex"`
	FirstName   string
	LastName    string
	IsVerified  bool `gorm:"default:false"`
	VerifiedAt  time.Time
	Role        string `gorm:"default:'user'"`
}

// Admin represents an administrator in the system
type Admin struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Username  string         `gorm:"uniqueIndex;not null"`
	Password  string         `gorm:"not null"`
}

// OTP represents a one-time password for user verification
type OTP struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	PhoneNumber string
	Code        string
	ExpiresAt   time.Time
}

// PasetoToken represents a PASETO token for user authentication
type PasetoToken struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	UserID       uuid.UUID      `gorm:"type:uuid;"`
	AccessToken  string         `gorm:"type:text"`
	RefreshToken string         `gorm:"type:text"`
	ExpiresAt    time.Time
}
