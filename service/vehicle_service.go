package service

import (
	"context"

	"shareway/repository"
	"shareway/schemas"
	"shareway/util"

	"github.com/google/uuid"
)

type IVehicleService interface {
	GetVehicles(ctx context.Context, limit int, page int, input string) ([]schemas.Vehicle, error)
	RegisterVehicle(userID uuid.UUID, vehicleID uuid.UUID, licensePlate string, caVet string) error
	LicensePlateExists(licensePlate string) (bool, error)
	CaVetExists(caVet string) (bool, error)
	GetVehicleFromID(vehicleID uuid.UUID) (schemas.VehicleDetail, error)
	GetAllVehiclesFromUserID(userID uuid.UUID) ([]schemas.VehicleDetail, error)
	GetTotalVehiclesForUser(userID uuid.UUID) (int64, error)
	GetVehiclesForUser(userID uuid.UUID) ([]schemas.VehicleDetail, error)
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

func (s *VehicleService) GetVehicles(ctx context.Context, limit int, page int, input string) ([]schemas.Vehicle, error) {
	vehicles, err := s.repo.GetVehicles(ctx, limit, page, input)
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

func (s *VehicleService) GetAllVehiclesFromUserID(userID uuid.UUID) ([]schemas.VehicleDetail, error) {
	return s.repo.GetAllVehiclesFromUserID(userID)
}

func (s *VehicleService) GetTotalVehiclesForUser(userID uuid.UUID) (int64, error) {
	return s.repo.GetTotalVehiclesForUser(userID)
}

func (s *VehicleService) GetVehiclesForUser(userID uuid.UUID) ([]schemas.VehicleDetail, error) {
	return s.repo.GetVehiclesForUser(userID)
}

var _ IVehicleService = (*VehicleService)(nil)
