package repository

import (
	"errors"
	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IRideRepository interface {
	GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error)
	AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error)
	CreateRideTransaction(rideID uuid.UUID, Fare float64, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error)
}

type RideRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRideRepository(db *gorm.DB, redis *redis.Client) IRideRepository {
	return &RideRepository{db: db, redis: redis}
}

var (
	ErrRideOfferNotFound   = errors.New("ride offer not found")
	ErrRideRequestNotFound = errors.New("ride request not found")
)

// GetRideOfferByID fetches a ride offer by its ID
func (r *RideRepository) GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error) {
	var rideOffer migration.RideOffer
	err := r.db.Model(&migration.RideOffer{}).
		Select("*"). // Replace with specific fields if you don't need all
		Where("id = ?", rideOfferID).
		Take(&rideOffer).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rideOffer, ErrRideOfferNotFound
		}
		return rideOffer, err
	}

	return rideOffer, nil
}

// GetRideRequestByID fetches a ride request by its ID
func (r *RideRepository) GetRideRequestByID(rideRequestID uuid.UUID) (migration.RideRequest, error) {
	var rideRequest migration.RideRequest
	err := r.db.Model(&migration.RideRequest{}).
		Select("*"). // Replace with specific fields if you don't need all
		Where("id = ?", rideRequestID).
		Take(&rideRequest).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rideRequest, ErrRideRequestNotFound
		}
		return rideRequest, err
	}

	return rideRequest, nil
}

// AcceptGiveRideRequest accepts a give ride request
func (r *RideRepository) AcceptRideRequest(rideOfferID, rideRequestID, vehicleID uuid.UUID) (migration.Ride, error) {
	var ride migration.Ride

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the ride offer by ID with only necessary fields
		var rideOffer migration.RideOffer
		err := tx.Select("start_time, end_time, status, fare, start_address, end_address, encoded_polyline, distance, duration, start_latitude, start_longitude, end_latitude, end_longitude").
			Where("id = ?", rideOfferID).
			First(&rideOffer).Error
		if err != nil {
			return err
		}
		// Check if the ride offer is already matched
		if rideOffer.Status == "matched" {
			return errors.New("ride offer is already matched")
		}

		// Get the ride request by ID with only necessary fields
		var rideRequest migration.RideRequest
		err = tx.Select("start_time, end_time, status, start_address, end_address, start_latitude, start_longitude, end_latitude, end_longitude, encoded_polyline, distance, duration,").
			Where("id = ?", rideRequestID).
			First(&rideRequest).Error
		if err != nil {
			return err
		}
		// Check if the ride request is already matched
		if rideRequest.Status == "matched" {
			return errors.New("ride request is already matched")
		}

		// Create a new ride
		ride = migration.Ride{
			RideOfferID:     rideOfferID,
			RideRequestID:   rideRequestID,
			Status:          "scheduled",
			StartTime:       rideOffer.StartTime,
			EndTime:         rideOffer.EndTime,
			Fare:            rideOffer.Fare,
			StartAddress:    rideOffer.StartAddress,
			EndAddress:      rideOffer.EndAddress,
			EncodedPolyline: rideOffer.EncodedPolyline,
			Distance:        rideOffer.Distance,
			Duration:        rideOffer.Duration,
			StartLatitude:   rideOffer.StartLatitude,
			StartLongitude:  rideOffer.StartLongitude,
			EndLatitude:     rideOffer.EndLatitude,
			EndLongitude:    rideOffer.EndLongitude,
			VehicleID:       vehicleID,
		}

		// Create the ride
		if err := tx.Create(&ride).Error; err != nil {
			return err
		}

		// Update ride offer status
		if err := tx.Model(&migration.RideOffer{}).Where("id = ?", rideOfferID).Update("status", "matched").Error; err != nil {
			return err
		}

		// Update ride request status
		if err := tx.Model(&migration.RideRequest{}).Where("id = ?", rideRequestID).Update("status", "matched").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Ride{}, err
	}

	return ride, nil
}

// CreateRideTransaction creates a transaction for a ride
func (r *RideRepository) CreateRideTransaction(rideID uuid.UUID, Fare float64, payerID uuid.UUID, receiverID uuid.UUID) (migration.Transaction, error) {
	var transaction migration.Transaction

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Create a new transaction
		transaction = migration.Transaction{
			RideID:        rideID,
			Amount:        Fare,
			Status:        "pending",
			PaymentMethod: "cash",
			PayerID:       payerID,
			ReceiverID:    receiverID,
		}

		// Create the transaction
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return migration.Transaction{}, err
	}

	return transaction, nil
}

// Make sure the RideRepository implements the IRideRepository interface
var _ IRideRepository = (*RideRepository)(nil)
