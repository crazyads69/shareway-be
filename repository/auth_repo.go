package repository

import (
	"gorm.io/gorm"
)

// Define UserRepository
type AuthRepository struct {
	DB *gorm.DB
}

// NewAuthRepository creates a new instance of AuthRepository
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{
		DB: db,
	}
}

// Define the interface for the AuthRepository
type IAuthRepository interface {
	// Define the methods for the AuthRepository
}

// Implement the methods for the AuthRepository
