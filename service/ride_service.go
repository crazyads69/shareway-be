package service

import (
	"shareway/infra/db/migration"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"

	"github.com/google/uuid"
)

type RideService struct {
	repo repository.IRideRepository
	hub  *ws.Hub
	cfg  util.Config
}

type IRideService interface {
	GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error)
	AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error)
	CreateRideTransaction(rideID uuid.UUID, Fare float64, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error)
	StartRide(req schemas.StartRideRequest, userID uuid.UUID) (migration.Ride, error)
	// EndRide(req schemas.EndRideRequest, userID uuid.UUID) (migration.Ride, error)
}

func NewRideService(repo repository.IRideRepository, hub *ws.Hub, cfg util.Config) IRideService {
	return &RideService{
		repo: repo,
		hub:  hub,
		cfg:  cfg,
	}
}

// GetRideOfferByID fetches a ride offer by its ID
func (s *RideService) GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error) {
	return s.repo.GetRideOfferByID(rideOfferID)
}

// GetRideRequestByID fetches a ride request by its ID
func (s *RideService) GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error) {
	return s.repo.GetRideRequestByID(rideRequestID)
}

// AcceptGiveRideRequest accepts a give ride request
func (s *RideService) AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error) {
	return s.repo.AcceptRideRequest(rideOfferID, rideRequestID, vehicleID)
}

// CreateRideTransaction creates a transaction for a ride
func (s *RideService) CreateRideTransaction(rideID uuid.UUID, Fare float64, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error) {
	return s.repo.CreateRideTransaction(rideID, Fare, payerID, receiverID)
}

// StartRide starts a ride
func (s *RideService) StartRide(req schemas.StartRideRequest, userID uuid.UUID) (migration.Ride, error) {
	return s.repo.StartRide(req, userID)
}

// EndRide ends a ride
// func (s *RideService) EndRide(req schemas.EndRideRequest, userID uuid.UUID) (migration.Ride, error) {
// 	return s.repo.EndRide(req, userID)
// }

// Make sure the RideService implements the IRideService interface
var _ IRideService = (*RideService)(nil)
