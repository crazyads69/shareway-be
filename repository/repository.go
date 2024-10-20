package repository

import (
	"shareway/util"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RepositoryContainer holds all the repositories
type RepositoryContainer struct {
	AuthRepository IAuthRepository
	MapsRepository IMapsRepository
	OTPRepository  IOTPRepository
	// Add other repositories here as needed
}

// RepositoryFactory is responsible for creating and initializing repositories
type RepositoryFactory struct {
	db          *gorm.DB
	redisClient *redis.Client
	cfg         util.Config
}

// NewRepositoryFactory creates a new RepositoryFactory
func NewRepositoryFactory(db *gorm.DB, redisClient *redis.Client, cfg util.Config) *RepositoryFactory {
	return &RepositoryFactory{
		db:          db,
		redisClient: redisClient,
		cfg:         cfg,
	}
}

// CreateRepositories initializes and returns all repositories
func (f *RepositoryFactory) CreateRepositories() *RepositoryContainer {
	return &RepositoryContainer{
		AuthRepository: f.createAuthRepository(),
		MapsRepository: f.createMapsRepository(),
		OTPRepository:  f.createOTPRepository(),
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

// createOTPRepository initializes and returns the OTP repository
func (f *RepositoryFactory) createOTPRepository() IOTPRepository {
	return NewOTPRepository(f.redisClient, f.cfg)
}

// Add methods for creating other repositories as needed
