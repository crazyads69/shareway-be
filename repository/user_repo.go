package repository

import (
	"gorm.io/gorm"
)

// Define UserRepository
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}
