package service

import (
	"shareway/infra/fpt"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/util"
	"shareway/util/token"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContainer struct {
	OTPService     IOTPService
	UserService    IUsersService
	MapService     IMapService
	VehicleService IVehicleService
	RideService    IRideService
}

type ServiceFactory struct {
	repos     *repository.RepositoryContainer
	cfg       util.Config
	fptReader *fpt.FPTReader
	encryptor util.IEncryptor
	maker     *token.PasetoMaker
	redis     *redis.Client
	hub       *ws.Hub
}

func NewServiceFactory(db *gorm.DB, cfg util.Config, token *token.PasetoMaker, redisClient *redis.Client, hub *ws.Hub) *ServiceFactory {
	repoFactory := repository.NewRepositoryFactory(db, redisClient, cfg)
	repos := repoFactory.CreateRepositories()

	// Initialize FPT reader
	fptReader := fpt.NewFPTReader(cfg)
	// Initialize encryptor
	encryptor := util.NewEncryptor(cfg)

	return &ServiceFactory{
		repos:     repos,
		cfg:       cfg,
		fptReader: fptReader,
		encryptor: encryptor,
		maker:     token,
		redis:     redisClient,
		hub:       hub,
	}
}

func (f *ServiceFactory) CreateServices() *ServiceContainer {
	return &ServiceContainer{
		OTPService:     f.createOTPService(),
		UserService:    f.createUserService(),
		MapService:     f.createMapsService(),
		VehicleService: f.createVehicleService(),
		RideService:    f.createRideService(),
	}
}

func (f *ServiceFactory) createOTPService() IOTPService {
	return NewOTPService(f.cfg, f.repos.OTPRepository)
}

func (f *ServiceFactory) createUserService() IUsersService {
	return NewUsersService(f.repos.AuthRepository, f.encryptor, f.fptReader, f.maker, f.cfg)
}

func (f *ServiceFactory) createMapsService() IMapService {
	return NewMapService(f.repos.MapsRepository, f.cfg, f.redis)
}

func (f *ServiceFactory) createVehicleService() IVehicleService {
	return NewVehicleService(f.repos.VehicleRepository, f.cfg)
}

func (f *ServiceFactory) createRideService() IRideService {
	return NewRideService(f.repos.RideRepository, f.hub, f.cfg)
}
