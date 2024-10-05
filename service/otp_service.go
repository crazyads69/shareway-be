package service

import (
	"shareway/repository"
)

type OTPService struct {
	repo repository.IOTPRepository
}
