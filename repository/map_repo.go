package repository

import (
	"errors"
	"fmt"
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type IMapsRepository interface {
	CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error)
	CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error)
	GetRideOfferDetails(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestDetails(rideRequestID uuid.UUID) (migration.RideRequest, error)
	SuggestRideRequests(userID uuid.UUID, rideOfferID uuid.UUID) ([]migration.RideRequest, error)
}

type MapsRepository struct {
	db *gorm.DB
}

func NewMapsRepository(db *gorm.DB) IMapsRepository {
	return &MapsRepository{db: db}
}

func (r *MapsRepository) CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error) {
	// Validate route data
	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
		log.Error().Msg("Invalid route data")
		return uuid.Nil, errors.New("invalid route data")
	}

	// Check if the vehicle exists and belongs to the user
	var vehicle migration.Vehicle
	if err := r.db.Preload("VehicleType").Where("id = ? AND user_id = ?", vehicleID, userID).First(&vehicle).Error; err != nil {
		return uuid.Nil, err
	}

	firstRoute := route.Routes[0]
	firstLeg := firstRoute.Legs[0]
	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

	// Calculate total distance and duration
	totalDistance, totalDuration := 0, 0
	for _, leg := range firstRoute.Legs {
		totalDistance += leg.Distance.Value
		totalDuration += leg.Duration.Value
	}
	totalDistance /= 1000 // Convert to kilometers

	// Calculate fare
	var fuelPrice migration.FuelPrice
	if err := r.db.Where("fuel_type = ?", "Xăng RON 95-III").First(&fuelPrice).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to fetch fuel price: %w", err)
	}

	fuelConsumed := vehicle.VehicleType.FuelConsumed / 100
	fare := fuelConsumed * fuelPrice.Price * float64(totalDistance)

	// Create ride offer
	rideOffer := migration.RideOffer{
		UserID:                 userID,
		StartLatitude:          firstLeg.Start_location.Lat,
		StartLongitude:         firstLeg.Start_location.Lng,
		EndLatitude:            lastLeg.End_location.Lat,
		EndLongitude:           lastLeg.End_location.Lng,
		EncodedPolyline:        firstRoute.Overview_polyline.Points,
		DriverCurrentLatitude:  currentLocation.Lat,
		DriverCurrentLongitude: currentLocation.Lng,
		StartAddress:           firstLeg.Start_address,
		EndAddress:             lastLeg.End_address,
		Distance:               float64(totalDistance),
		Duration:               totalDuration,
		Status:                 "created",
		StartTime:              startTime,
		EndTime:                startTime.Add(time.Duration(totalDuration) * time.Second),
		VehicleID:              vehicleID,
		Fare:                   fare,
	}

	// Use a transaction to ensure atomicity
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rideOffer).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create ride offer")
			return err
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return rideOffer.ID, nil
}
func (r *MapsRepository) CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error) {
	// Validate route data
	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
		return uuid.Nil, errors.New("invalid route data")
	}

	firstRoute := route.Routes[0]
	firstLeg := firstRoute.Legs[0]
	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

	// Calculate total distance and duration
	totalDistance, totalDuration := 0, 0
	for _, leg := range firstRoute.Legs {
		totalDistance += leg.Distance.Value
		totalDuration += leg.Duration.Value
	}
	totalDistance /= 1000 // Convert to kilometers

	// Create ride request
	rideRequest := &migration.RideRequest{
		UserID:                userID,
		StartLatitude:         firstLeg.Start_location.Lat,
		StartLongitude:        firstLeg.Start_location.Lng,
		EndLatitude:           lastLeg.End_location.Lat,
		EndLongitude:          lastLeg.End_location.Lng,
		RiderCurrentLatitude:  currentLocation.Lat,
		RiderCurrentLongitude: currentLocation.Lng,
		StartAddress:          firstLeg.Start_address,
		EndAddress:            lastLeg.End_address,
		Status:                "created",
		EncodedPolyline:       firstRoute.Overview_polyline.Points,
		Distance:              float64(totalDistance),
		Duration:              totalDuration,
		StartTime:             startTime,
		EndTime:               startTime.Add(time.Duration(totalDuration) * time.Second),
	}

	if err := r.db.Create(rideRequest).Error; err != nil {
		return uuid.Nil, err
	}
	return rideRequest.ID, nil
}

func (r *MapsRepository) GetRideOfferDetails(rideOfferID uuid.UUID) (migration.RideOffer, error) {
	rideOffer := migration.RideOffer{}
	if err := r.db.Preload("Vehicle").First(&rideOffer, rideOfferID).Error; err != nil {
		return migration.RideOffer{}, err
	}
	return rideOffer, nil
}

func (r *MapsRepository) GetRideRequestDetails(rideRequestID uuid.UUID) (migration.RideRequest, error) {
	rideRequest := migration.RideRequest{}
	if err := r.db.First(&rideRequest, rideRequestID).Error; err != nil {
		return migration.RideRequest{}, err
	}
	return rideRequest, nil
}

func (r *MapsRepository) SuggestRideRequests(userID uuid.UUID, rideOfferID uuid.UUID) ([]migration.RideRequest, error) {
	// Fetch the ride offer details
	rideOffer, err := r.GetRideOfferDetails(rideOfferID)
	if err != nil {
		return nil, err
	}

	// Fetch the ride requests that have status "created"
	var rideRequests []migration.RideRequest
	if err := r.db.Where("status = ?", "created").Find(&rideRequests).Error; err != nil {
		return nil, err
	}

	var filteredRideRequests []migration.RideRequest
	offerPolyline := helper.DecodePolyline(rideOffer.EncodedPolyline)
	const maxDistance = 2.0 // km

	for _, rideRequest := range rideRequests {
		requestPolyline := helper.DecodePolyline(rideRequest.EncodedPolyline)

		if rideRequest.UserID != userID && helper.IsRouteMatching(offerPolyline, requestPolyline, maxDistance) &&
			helper.IsTimeOverlap(rideOffer, rideRequest) {
			filteredRideRequests = append(filteredRideRequests, rideRequest)
		}
	}

	return filteredRideRequests, nil
}

// Make sure to implement the IMapsRepository interface
var _ IMapsRepository = (*MapsRepository)(nil)
