package repository

import (
	"gorm.io/gorm"
)

// RepositoryContainer holds all the repositories
type RepositoryContainer struct {
	AuthRepository IAuthRepository
	MapsRepository IMapsRepository
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
		MapsRepository: f.createMapsRepository(),
		// Initialize other repositories here
	}
}

// createAuthRepository initializes and returns the Auth repository
func (f *RepositoryFactory) createAuthRepository() IAuthRepository {
	return NewAuthRepository(f.db)
}

// createMapsRepository initializes and returns the Maps repository
func (f *RepositoryFactory) createMapsRepository() IMapsRepository {
	return NewMapsRepository(f.db)
}

// Add methods for creating other repositories as needed
