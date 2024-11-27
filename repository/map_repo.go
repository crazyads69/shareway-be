package repository

import (
	"errors"
	"fmt"
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"shareway/util/polyline"
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
	GetRideByID(rideID uuid.UUID) (migration.Ride, error)
	GetAllWaypoints(rideOfferID uuid.UUID) ([]migration.Waypoint, error)
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
		// Check if the vehicle exists and belongs to the user
		var vehicle migration.Vehicle
		if err := r.db.Preload("VehicleType").Where("id = ? AND user_id = ?", vehicleID, userID).First(&vehicle).Error; err != nil {
			return err
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

		// Check for overlapping ride requests
		var existingRideRequestCount int64
		err = tx.Model(&migration.RideRequest{}).
			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
				userID, startTime, endTime, startTime, endTime, startTime, endTime).
			Count(&existingRideRequestCount).Error
		if err != nil {
			return fmt.Errorf("error checking for existing ride requests: %w", err)
		}
		if existingRideRequestCount > 0 {
			return errors.New("ride request already exists for the user in that time frame")
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
		// Decode the polyline and get the list of coordinates
		decodePolyline := helper.DecodePolyline(firstRoute.Overview_polyline.Points)
		startLocation := schemas.Point{
			Lat: firstLeg.Start_location.Lat,
			Lng: firstLeg.Start_location.Lng,
		}
		endLocation := schemas.Point{
			Lat: lastLeg.End_location.Lat,
			Lng: lastLeg.End_location.Lng,
		}

		// Get the correct start and end locations on the route
		newStartLocaton, newEndLocation := helper.FindClosestPoints(decodePolyline, startLocation, endLocation)

		// Create ride offer
		// rideOffer := migration.RideOffer{
		// 	UserID:                 userID,
		// 	StartLatitude:          firstLeg.Start_location.Lat,
		// 	StartLongitude:         firstLeg.Start_location.Lng,
		// 	EndLatitude:            lastLeg.End_location.Lat,
		// 	EndLongitude:           lastLeg.End_location.Lng,
		// 	EncodedPolyline:        polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
		// 	DriverCurrentLatitude:  currentLocation.Lat,
		// 	DriverCurrentLongitude: currentLocation.Lng,
		// 	StartAddress:           firstLeg.Start_address,
		// 	EndAddress:             lastLeg.End_address,
		// 	Distance:               float64(totalDistance),
		// 	Duration:               totalDuration,
		// 	Status:                 "created",
		// 	StartTime:              startTime,
		// 	EndTime:                endTime,
		// 	VehicleID:              vehicleID,
		// 	Fare:                   fare,
		// }

		rideOffer := migration.RideOffer{
			UserID:                 userID,
			StartLatitude:          newStartLocaton.Lat,
			StartLongitude:         newStartLocaton.Lng,
			EndLatitude:            newEndLocation.Lat,
			EndLongitude:           newEndLocation.Lng,
			EncodedPolyline:        polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
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

		// Only create waypoints if there are more than 2 legs
		if len(firstRoute.Legs) < 2 {
			// no need to create waypoints if there are only 1
			return nil
		}

		// Create waypoints
		newWaypoints := make([]migration.Waypoint, 0, len(route.Geocoded_waypoints)-2) // Exclude the start and end locations
		// TODO: Find closest points on the route for the start and end locations
		// only need to store waypoints that are not the start or end locations
		// And the waypoints is end location of the previous leg and start location of the next leg
		for i, leg := range firstRoute.Legs {
			if i == 0 {
				continue
			}
			if i == len(firstRoute.Legs)-1 {
				break
			}

			legEnd := leg.End_location

			waypoint := migration.Waypoint{
				RideOfferID:   rideOfferID,
				Latitude:      legEnd.Lat,
				Longitude:     legEnd.Lng,
				WaypointOrder: i,
			}
			newWaypoints = append(newWaypoints, waypoint)
		}
		if err := tx.Create(&newWaypoints).Error; err != nil {
			return fmt.Errorf("failed to create waypoints: %w", err)
		}

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

		// Check if any ride offer exists for the user that overlaps with the new time frame
		var existingRideOfferCount int64
		err = tx.Model(&migration.RideOffer{}).
			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
				userID, startTime, endTime, startTime, endTime, startTime, endTime).
			Count(&existingRideOfferCount).Error
		if err != nil {
			return fmt.Errorf("error checking for existing ride offers: %w", err)
		}
		if existingRideOfferCount > 0 {
			return errors.New("ride offer already exists for the user in that time frame")
		}

		decodePolyline := helper.DecodePolyline(firstRoute.Overview_polyline.Points)
		startLocation := schemas.Point{
			Lat: firstLeg.Start_location.Lat,
			Lng: firstLeg.Start_location.Lng,
		}
		endLocation := schemas.Point{
			Lat: lastLeg.End_location.Lat,
			Lng: lastLeg.End_location.Lng,
		}

		// Get the correct start and end locations on the route
		newStartLocaton, newEndLocation := helper.FindClosestPoints(decodePolyline, startLocation, endLocation)

		// Create ride request
		// rideRequest := migration.RideRequest{
		// 	UserID:                userID,
		// 	StartLatitude:         firstLeg.Start_location.Lat,
		// 	StartLongitude:        firstLeg.Start_location.Lng,
		// 	EndLatitude:           lastLeg.End_location.Lat,
		// 	EndLongitude:          lastLeg.End_location.Lng,
		// 	RiderCurrentLatitude:  currentLocation.Lat,
		// 	RiderCurrentLongitude: currentLocation.Lng,
		// 	StartAddress:          firstLeg.Start_address,
		// 	EndAddress:            lastLeg.End_address,
		// 	Status:                "created",
		// 	EncodedPolyline:       polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
		// 	Distance:              float64(totalDistance),
		// 	Duration:              totalDuration,
		// 	StartTime:             startTime,
		// 	EndTime:               endTime,
		// }

		rideRequest := migration.RideRequest{
			UserID:                userID,
			StartLatitude:         newStartLocaton.Lat,
			StartLongitude:        newStartLocaton.Lng,
			EndLatitude:           newEndLocation.Lat,
			EndLongitude:          newEndLocation.Lng,
			RiderCurrentLatitude:  currentLocation.Lat,
			RiderCurrentLongitude: currentLocation.Lng,
			StartAddress:          firstLeg.Start_address,
			EndAddress:            lastLeg.End_address,
			Status:                "created",
			EncodedPolyline:       polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
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
	offerPolyline := helper.DecodePolyline(string(rideOffer.EncodedPolyline))
	const maxDistance = 2.0 // km

	for _, rideRequest := range rideRequests {
		requestPolyline := helper.DecodePolyline(string(rideRequest.EncodedPolyline))

		if rideRequest.UserID != userID && helper.IsMatchRoute(offerPolyline, requestPolyline) &&
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
	requestPolyline := helper.DecodePolyline(string(rideRequest.EncodedPolyline))
	const maxDistance = 2.0 // km

	for _, rideOffer := range rideOffers {
		offerPolyline := helper.DecodePolyline(string(rideOffer.EncodedPolyline))

		if rideOffer.UserID != userID && helper.IsMatchRoute(offerPolyline, requestPolyline) &&
			helper.IsTimeOverlap(rideOffer, rideRequest) {
			filteredRideOffers = append(filteredRideOffers, rideOffer)
		}
	}

	return filteredRideOffers, nil
}

func (r *MapsRepository) GetRideByID(rideID uuid.UUID) (migration.Ride, error) {
	ride := migration.Ride{}
	if err := r.db.Preload("RideOffer").Preload("RideRequest").First(&ride, rideID).Error; err != nil {
		return migration.Ride{}, err
	}
	return ride, nil
}

func (r *MapsRepository) GetAllWaypoints(rideOfferID uuid.UUID) ([]migration.Waypoint, error) {
	var waypoints []migration.Waypoint
	err := r.db.Where("ride_offer_id = ?", rideOfferID).Order("waypoint_order").Find(&waypoints).Error

	if err != nil {
		return nil, err
	}

	return waypoints, nil
}

// Make sure to implement the IMapsRepository interface
var _ IMapsRepository = (*MapsRepository)(nil)
