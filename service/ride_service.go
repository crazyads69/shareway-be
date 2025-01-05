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
	CreateNewChatRoom(userID1, userID2 uuid.UUID) error
	GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error)
	GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error)
	GetTransactionByRideID(rideID uuid.UUID) (migration.Transaction, error)
	AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error)
	CreateRideTransaction(rideID uuid.UUID, Fare int64, paymentMethod string, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error)
	StartRide(req schemas.StartRideRequest, userID uuid.UUID) (migration.Ride, error)
	EndRide(req schemas.EndRideRequest, userID uuid.UUID) (migration.Ride, error)
	UpdateRideLocation(req schemas.UpdateRideLocationRequest, userID uuid.UUID) (migration.Ride, error)
	CancelRide(req schemas.CancelRideRequest, userID uuid.UUID) (migration.Ride, error)
	GetAllPendingRide(userID uuid.UUID) ([]migration.RideOffer, []migration.RideRequest, error)
	GetRideByID(rideID uuid.UUID) (migration.Ride, error)
	RatingRideHitcher(req schemas.RatingRideHitcherRequest, userID uuid.UUID) error
	RatingRideDriver(req schemas.RatingRideDriverRequest, userID uuid.UUID) error
	GetRideHistory(userID uuid.UUID) ([]migration.Ride, error)
	GetTotalRidesForUser(userID uuid.UUID) (int64, error)
	GetTotalRidesForVehicle(vehicleID uuid.UUID) (int64, error)
	GetScheduledAndOngoingRide(userID uuid.UUID) ([]migration.Ride, error)
}

func NewRideService(repo repository.IRideRepository, hub *ws.Hub, cfg util.Config) IRideService {
	return &RideService{
		repo: repo,
		hub:  hub,
		cfg:  cfg,
	}
}

// CreateNewChatRoom creates a new chat room between two users
func (s *RideService) CreateNewChatRoom(userID1, userID2 uuid.UUID) error {
	return s.repo.CreateNewChatRoom(userID1, userID2)
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
func (s *RideService) CreateRideTransaction(rideID uuid.UUID, Fare int64, paymentMethod string, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error) {
	return s.repo.CreateRideTransaction(rideID, Fare, paymentMethod, payerID, receiverID)
}

// StartRide starts a ride
func (s *RideService) StartRide(req schemas.StartRideRequest, userID uuid.UUID) (migration.Ride, error) {
	return s.repo.StartRide(req, userID)
}

// GetTransactionByID fetches a transaction by its ID
func (s *RideService) GetTransactionByRideID(rideID uuid.UUID) (migration.Transaction, error) {
	return s.repo.GetTransactionByRideID(rideID)
}

// EndRide ends a ride
func (s *RideService) EndRide(req schemas.EndRideRequest, userID uuid.UUID) (migration.Ride, error) {
	return s.repo.EndRide(req, userID)
}

// UpdateRideLocation updates the location of a ride
func (s *RideService) UpdateRideLocation(req schemas.UpdateRideLocationRequest, userID uuid.UUID) (migration.Ride, error) {
	return s.repo.UpdateRideLocation(req, userID)
}

// CancelRideByDriver cancels a ride by the driver
func (s *RideService) CancelRide(req schemas.CancelRideRequest, userID uuid.UUID) (migration.Ride, error) {
	return s.repo.CancelRide(req, userID)
}

func (s *RideService) GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error) {
	return s.repo.GetChatRoomByUserIDs(userID1, userID2)
}

func (s *RideService) GetAllPendingRide(userID uuid.UUID) ([]migration.RideOffer, []migration.RideRequest, error) {
	return s.repo.GetAllPendingRide(userID)
}

func (s *RideService) GetRideByID(rideID uuid.UUID) (migration.Ride, error) {
	return s.repo.GetRideByID(rideID)
}

func (s *RideService) RatingRideHitcher(req schemas.RatingRideHitcherRequest, userID uuid.UUID) error {
	return s.repo.RatingRideHitcher(req, userID)
}

func (s *RideService) RatingRideDriver(req schemas.RatingRideDriverRequest, userID uuid.UUID) error {
	return s.repo.RatingRideDriver(req, userID)
}

func (s *RideService) GetRideHistory(userID uuid.UUID) ([]migration.Ride, error) {
	return s.repo.GetRideHistory(userID)
}

func (s *RideService) GetTotalRidesForUser(userID uuid.UUID) (int64, error) {
	return s.repo.GetTotalRidesForUser(userID)
}

func (s *RideService) GetTotalRidesForVehicle(vehicleID uuid.UUID) (int64, error) {
	return s.repo.GetTotalRidesForVehicle(vehicleID)
}

func (s *RideService) GetScheduledAndOngoingRide(userID uuid.UUID) ([]migration.Ride, error) {
	return s.repo.GetScheduledAndOngoingRide(userID)
}

// Make sure the RideService implements the IRideService interface
var _ IRideService = (*RideService)(nil)
