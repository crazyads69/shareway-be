package service

import (
	"shareway/repository"
	"shareway/util"

	"gorm.io/gorm"
)

type ServiceContainer struct {
	OTPService  IOTPService
	UserService IUsersService
}

type ServiceFactory struct {
	repos *repository.RepositoryContainer
	cfg   util.Config
}

func NewServiceFactory(db *gorm.DB, cfg util.Config) *ServiceFactory {
	repoFactory := repository.NewRepositoryFactory(db)
	repos := repoFactory.CreateRepositories()

	return &ServiceFactory{
		repos: repos,
		cfg:   cfg,
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
	return NewUsersService(f.repos.AuthRepository)
}
