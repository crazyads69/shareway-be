package repository

import (
	"errors"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMapsRepository interface {
	CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error)
	CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error)
	CalculateFareForRideOffer(rideOfferID uuid.UUID, vehicleID uuid.UUID) (float64, error)
	GetRideOfferDetails(rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestDetails(rideRequestID uuid.UUID) (migration.RideRequest, error)
}

type MapsRepository struct {
	db *gorm.DB
}

func NewMapsRepository(db *gorm.DB) IMapsRepository {
	return &MapsRepository{db: db}
}

func (r *MapsRepository) CalculateFareForRideOffer(rideOfferID uuid.UUID, vehicleID uuid.UUID) (float64, error) {
	// Optimize by using a single query to fetch all required data
	var result struct {
		FuelConsumed float64
		Distance     int
		FuelPrice    float64
	}

	err := r.db.Table("vehicles").
		Select("vehicle_types.fuel_consumed, ride_offers.distance, fuel_prices.price").
		Joins("JOIN vehicle_types ON vehicles.vehicle_type_id = vehicle_types.id").
		Joins("JOIN ride_offers ON ride_offers.id = ?", rideOfferID).
		Joins("JOIN fuel_prices ON fuel_prices.fuel_type = ?", "Xăng RON 95-III").
		Where("vehicles.id = ?", vehicleID).
		First(&result).Error

	if err != nil {
		return 0, err
	}

	// Calculate and return the fare
	return result.FuelConsumed / 100 * float64(result.Distance) * result.FuelPrice, nil
}

func (r *MapsRepository) CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error) {
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
	}

	// Use a transaction to ensure atomicity
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rideOffer).Error; err != nil {
			return err
		}

		fare, err := r.CalculateFareForRideOffer(rideOffer.ID, vehicleID)
		if err != nil {
			return err
		}

		return tx.Model(&rideOffer).Update("fare", fare).Error
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

// Make sure to implement the IMapsRepository interface
var _ IMapsRepository = (*MapsRepository)(nil)
