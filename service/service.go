package service

import (
	"shareway/repository"
	"shareway/util"

	"gorm.io/gorm"
)

type Service struct {
	OtpService  *OtpService
	UserService *UsersService
}

func NewService(db *gorm.DB, cfg util.Config) *Service {
	// init repositories
	userRepo := repository.NewAuthRepository(db)
	// init services
	userService := NewUsersService(userRepo)
	otpService := NewOTPService(cfg)
	return &Service{
		OtpService:  otpService,
		UserService: userService,
	}
}
