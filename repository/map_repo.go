package repository

import (
	"errors"
	"fmt"
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMapsRepository interface {
	CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error)
	CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error)
	GetRideOfferDetails(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestDetails(rideRequestID uuid.UUID) (migration.RideRequest, error)
	SuggestRideRequests(userID uuid.UUID, rideOfferID uuid.UUID) ([]migration.RideRequest, error)
	SuggestRideOffers(userID uuid.UUID, rideRequestID uuid.UUID) ([]migration.RideOffer, error)
}

type MapsRepository struct {
	db *gorm.DB
}

func NewMapsRepository(db *gorm.DB) IMapsRepository {
	return &MapsRepository{db: db}
}
func (r *MapsRepository) CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error) {
	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
		return uuid.Nil, errors.New("invalid route data")
	}

	firstRoute := route.Routes[0]
	firstLeg := firstRoute.Legs[0]
	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

	totalDistance, totalDuration := 0, 0
	for _, leg := range firstRoute.Legs {
		totalDistance += leg.Distance.Value
		totalDuration += leg.Duration.Value
	}
	totalDistance /= 1000 // Convert to kilometers

	endTime := startTime.Add(time.Duration(totalDuration) * time.Second)

	var rideOfferID uuid.UUID
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Check if the vehicle exists and belongs to the user
		var vehicle struct {
			FuelConsumed float64
		}
		if err := tx.Table("vehicles").
			Select("vehicle_types.fuel_consumed").
			Joins("JOIN vehicle_types ON vehicles.vehicle_type_id = vehicle_types.id").
			Where("vehicles.id = ? AND vehicles.user_id = ?", vehicleID, userID).
			First(&vehicle).Error; err != nil {
			return fmt.Errorf("failed to fetch vehicle: %w", err)
		}

		// Check for overlapping ride offers
		var existingRideOfferCount int64
		err := tx.Model(&migration.RideOffer{}).
			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
				userID, startTime, endTime, startTime, endTime, startTime, endTime).
			Count(&existingRideOfferCount).Error
		if err != nil {
			return fmt.Errorf("error checking for existing ride offers: %w", err)
		}
		if existingRideOfferCount > 0 {
			return errors.New("ride offer already exists for the user in that time frame")
		}

		// Fetch fuel price
		var fuelPrice float64
		if err := tx.Model(&migration.FuelPrice{}).
			Select("price").
			Where("fuel_type = ?", "Xăng RON 95-III").
			First(&fuelPrice).Error; err != nil {
			return fmt.Errorf("failed to fetch fuel price: %w", err)
		}

		fare := (vehicle.FuelConsumed / 100) * fuelPrice * float64(totalDistance)

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
			EndTime:                endTime,
			VehicleID:              vehicleID,
			Fare:                   fare,
		}

		if err := tx.Create(&rideOffer).Error; err != nil {
			return fmt.Errorf("failed to create ride offer: %w", err)
		}

		rideOfferID = rideOffer.ID
		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return rideOfferID, nil
}
func (r *MapsRepository) CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error) {
	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
		return uuid.Nil, errors.New("invalid route data")
	}

	firstRoute := route.Routes[0]
	firstLeg := firstRoute.Legs[0]
	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

	totalDistance, totalDuration := 0, 0
	for _, leg := range firstRoute.Legs {
		totalDistance += leg.Distance.Value
		totalDuration += leg.Duration.Value
	}
	totalDistance /= 1000 // Convert to kilometers

	endTime := startTime.Add(time.Duration(totalDuration) * time.Second)

	var rideRequestID uuid.UUID
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Check if any ride request exists for the user that overlaps with the new time frame
		var existingRideRequestCount int64
		err := tx.Model(&migration.RideRequest{}).
			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
				userID, startTime, endTime, startTime, endTime, startTime, endTime).
			Count(&existingRideRequestCount).Error
		if err != nil {
			return fmt.Errorf("error checking for existing ride requests: %w", err)
		}
		if existingRideRequestCount > 0 {
			return errors.New("ride request already exists for the user in that time frame")
		}

		// Create ride request
		rideRequest := migration.RideRequest{
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
			EndTime:               endTime,
		}

		if err := tx.Create(&rideRequest).Error; err != nil {
			return fmt.Errorf("error creating ride request: %w", err)
		}

		rideRequestID = rideRequest.ID
		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return rideRequestID, nil
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

// SuggestRideOffers suggests ride offers that match the given ride request
func (r *MapsRepository) SuggestRideOffers(userID uuid.UUID, rideRequestID uuid.UUID) ([]migration.RideOffer, error) {
	// Fetch the ride request details
	rideRequest, err := r.GetRideRequestDetails(rideRequestID)
	if err != nil {
		return nil, err
	}

	// Fetch the ride offers that have status "created"
	var rideOffers []migration.RideOffer
	if err := r.db.Where("status = ?", "created").Find(&rideOffers).Error; err != nil {
		return nil, err
	}

	var filteredRideOffers []migration.RideOffer
	requestPolyline := helper.DecodePolyline(rideRequest.EncodedPolyline)
	const maxDistance = 2.0 // km

	for _, rideOffer := range rideOffers {
		offerPolyline := helper.DecodePolyline(rideOffer.EncodedPolyline)

		if rideOffer.UserID != userID && helper.IsRouteMatching(offerPolyline, requestPolyline, maxDistance) &&
			helper.IsTimeOverlap(rideOffer, rideRequest) {
			filteredRideOffers = append(filteredRideOffers, rideOffer)
		}
	}

	return filteredRideOffers, nil
}

// Make sure to implement the IMapsRepository interface
var _ IMapsRepository = (*MapsRepository)(nil)
