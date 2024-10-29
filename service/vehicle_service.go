package service

import (
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"

	"github.com/google/uuid"
)

type IVehicleService interface {
	GetVehicles() ([]schemas.Vehicle, error)
	RegisterVehicle(userID uuid.UUID, vehicleID uuid.UUID, licensePlate string, caVet string) error
	LicensePlateExists(licensePlate string) (bool, error)
	CaVetExists(caVet string) (bool, error)
	GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error)
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

func (s *VehicleService) RegisterVehicle(userID uuid.UUID, vehicleID uuid.UUID, licensePlate string, caVet string) error {
	return s.repo.RegisterVehicle(userID, vehicleID, licensePlate, caVet)
}

func (s *VehicleService) LicensePlateExists(licensePlate string) (bool, error) {
	return s.repo.LicensePlateExists(licensePlate)
}

func (s *VehicleService) CaVetExists(caVet string) (bool, error) {
	return s.repo.CaVetExists(caVet)
}

func (s *VehicleService) GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error) {
	return s.repo.GetVehicleFromID(vehicleID)
}
