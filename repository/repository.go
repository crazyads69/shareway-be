package repository

import (
	"shareway/util"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RepositoryContainer holds all the repositories
type RepositoryContainer struct {
	AuthRepository         IAuthRepository
	MapsRepository         IMapsRepository
	OTPRepository          IOTPRepository
	VehicleRepository      IVehicleRepository
	RideRepository         IRideRepository
	NotificationRepository INotificationRepository
	ChatRepository         IChatRepository
	AdminRepository        IAdminRepository
	PaymentRepository      IPaymentRepository
	IPNRepository          IIPNRepository
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
		AuthRepository:         f.createAuthRepository(),
		MapsRepository:         f.createMapsRepository(),
		OTPRepository:          f.createOTPRepository(),
		VehicleRepository:      f.createVehicleRepository(),
		RideRepository:         f.createRideRepository(),
		NotificationRepository: f.createNotificationRepository(),
		ChatRepository:         f.createChatRepository(),
		AdminRepository:        f.createAdminRepository(),
		PaymentRepository:      f.createPaymentRepository(),
		IPNRepository:          f.createIPNRepository(),
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

// createVehicleRepository initializes and returns the Vehicle repository
func (f *RepositoryFactory) createVehicleRepository() IVehicleRepository {
	return NewVehicleRepository(f.db, f.redisClient)
}

// createRideRepository initializes and returns the Ride repository
func (f *RepositoryFactory) createRideRepository() IRideRepository {
	return NewRideRepository(f.db, f.redisClient)
}

// createNotificationRepository initializes and returns the Notification repository
func (f *RepositoryFactory) createNotificationRepository() INotificationRepository {
	return NewNotificationRepository(f.db)
}

// createChatRepository initializes and returns the Chat repository
func (f *RepositoryFactory) createChatRepository() IChatRepository {
	return NewChatRepository(f.db, f.redisClient)
}

// createAdminRepository initializes and returns the Admin repository
func (f *RepositoryFactory) createAdminRepository() IAdminRepository {
	return NewAdminRepository(f.db, f.redisClient)
}

// createPaymentRepository initializes and returns the Payment repository
func (f *RepositoryFactory) createPaymentRepository() IPaymentRepository {
	return NewPaymentRepository(f.db, f.redisClient)
}

// createIPNRepository initializes and returns the IPN repository
func (f *RepositoryFactory) createIPNRepository() IIPNRepository {
	return NewIPNRepository(f.db, f.redisClient)
}

// Add methods for creating other repositories as needed
