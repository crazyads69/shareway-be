package service

import (
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
)

type IVehicleService interface {
	GetVehicles() ([]schemas.Vehicle, error)
}

type VehicleService struct {
	repo repository.IVehicleRepository
	cfg  util.Config
}

func NewVehicleService(repo repository.IVehicleRepository, cfg util.Config) IVehicleService {
	return &VehicleService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *VehicleService) GetVehicles() ([]schemas.Vehicle, error) {
	vehicles, err := s.repo.GetVehicles()
	if err != nil {
		return nil, err
	}

	return vehicles, nil
}
