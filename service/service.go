package service

import (
	"shareway/infra/fpt"
	"shareway/repository"
	"shareway/util"
	"shareway/util/token"

	"gorm.io/gorm"
)

type ServiceContainer struct {
	OTPService  IOTPService
	UserService IUsersService
}

type ServiceFactory struct {
	repos     *repository.RepositoryContainer
	cfg       util.Config
	fptReader *fpt.FPTReader
	encryptor util.IEncryptor
	maker     *token.PasetoMaker
}

func NewServiceFactory(db *gorm.DB, cfg util.Config, token *token.PasetoMaker) *ServiceFactory {
	repoFactory := repository.NewRepositoryFactory(db)
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
	}
}

func (f *ServiceFactory) CreateServices() *ServiceContainer {
	return &ServiceContainer{
		OTPService:  f.createOTPService(),
		UserService: f.createUserService(),
	}
}

func (f *ServiceFactory) createOTPService() IOTPService {
	return NewOTPService(f.cfg)
}

func (f *ServiceFactory) createUserService() IUsersService {
	return NewUsersService(f.repos.AuthRepository, f.encryptor, f.fptReader, f.maker, f.cfg)
}
