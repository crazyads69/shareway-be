package service

import (
	"shareway/infra/fpt"
	"shareway/repository"
	"shareway/util"
	"shareway/util/token"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContainer struct {
	OTPService  IOTPService
	UserService IUsersService
	MapsService IMapsService
}

type ServiceFactory struct {
	repos     *repository.RepositoryContainer
	cfg       util.Config
	fptReader *fpt.FPTReader
	encryptor util.IEncryptor
	maker     *token.PasetoMaker
	redis     *redis.Client
}

func NewServiceFactory(db *gorm.DB, cfg util.Config, token *token.PasetoMaker, redisClient *redis.Client) *ServiceFactory {
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
	}
}

func (f *ServiceFactory) CreateServices() *ServiceContainer {
	return &ServiceContainer{
		OTPService:  f.createOTPService(),
		UserService: f.createUserService(),
		MapsService: f.createMapsService(),
	}
}

func (f *ServiceFactory) createOTPService() IOTPService {
	return NewOTPService(f.cfg, f.repos.OTPRepository)
}

func (f *ServiceFactory) createUserService() IUsersService {
	return NewUsersService(f.repos.AuthRepository, f.encryptor, f.fptReader, f.maker, f.cfg)
}

func (f *ServiceFactory) createMapsService() IMapsService {
	return NewMapsService(f.repos.MapsRepository, f.cfg, f.redis)
}
