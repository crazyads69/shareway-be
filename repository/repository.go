package repository

import (
	"gorm.io/gorm"
)

// RepositoryContainer holds all the repositories
type RepositoryContainer struct {
	AuthRepository IAuthRepository
	// Add other repositories here as needed
}

// RepositoryFactory is responsible for creating and initializing repositories
type RepositoryFactory struct {
	db *gorm.DB
}

// NewRepositoryFactory creates a new RepositoryFactory
func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// CreateRepositories initializes and returns all repositories
func (f *RepositoryFactory) CreateRepositories() *RepositoryContainer {
	return &RepositoryContainer{
		AuthRepository: f.createAuthRepository(),
		// Initialize other repositories here
	}
}

// createAuthRepository initializes and returns the Auth repository
func (f *RepositoryFactory) createAuthRepository() IAuthRepository {
	return NewAuthRepository(f.db)
}

// Add methods for creating other repositories as needed
