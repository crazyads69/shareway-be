package migration

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define User struct
type User struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	PhoneNumber string         `gorm:"uniqueIndex;not null"`
	Email       string         `gorm:"uniqueIndex"`
	FirstName   string
	LastName    string
	Password    string
	IsVerified  bool `gorm:"default:false"`
	VerifiedAt  time.Time
	Role        string `gorm:"default:'user'"`
}

// Define Admin struct
type Admin struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Username  string         `gorm:"uniqueIndex;not null"`
	Password  string         `gorm:"not null"`
}

// Define OTP struct
type OTP struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	PhoneNumber string
	Code        string
	ExpiresAt   time.Time
}

// Define Paseto struct

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
