package repository

import (
	"errors"
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMapsRepository interface {
	CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point) (uuid.UUID, error)
	CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point) (uuid.UUID, error)
}

type MapsRepository struct {
	db *gorm.DB
}

func NewMapsRepository(db *gorm.DB) IMapsRepository {
	return &MapsRepository{db: db}
}

func (r *MapsRepository) CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point) (uuid.UUID, error) {
	if len(route.Routes) == 0 {
		return uuid.Nil, errors.New("no routes found in the response")
	}

	firstRoute := route.Routes[0]
	if len(firstRoute.Legs) == 0 {
		return uuid.Nil, errors.New("no legs found in the first route")
	}

	firstLeg := firstRoute.Legs[0]
	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

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
		Distance:               firstRoute.Legs[0].Distance.Value,
		Duration:               firstRoute.Legs[0].Duration.Value,
		Status:                 "created", // Initial status is "created"
	}

	if err := r.db.Create(&rideOffer).Error; err != nil {
		return uuid.Nil, err
	}

	return rideOffer.ID, nil
}

func (r *MapsRepository) CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point) (uuid.UUID, error) {
	if len(route.Routes) == 0 {
		return uuid.Nil, errors.New("no routes found")
	}

	firstRoute := route.Routes[0]
	if len(firstRoute.Legs) == 0 {
		return uuid.Nil, errors.New("no legs found in the route")
	}

	firstLeg := firstRoute.Legs[0]
	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

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
		Distance:              firstRoute.Legs[0].Distance.Value,
		Duration:              firstRoute.Legs[0].Duration.Value,
	}

	if err := r.db.Create(rideRequest).Error; err != nil {
		return uuid.Nil, err
	}
	return rideRequest.ID, nil
}
