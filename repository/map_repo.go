package repository

import (
	"errors"
	"fmt"
	"math"
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"shareway/util/polyline"
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
	log.Debug().
		Interface("route", route).
		Str("userID", userID.String()).
		Interface("currentLocation", currentLocation).
		Time("startTime", startTime).
		Str("vehicleID", vehicleID.String()).
		Msg("CreateGiveRide function called")

	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
		log.Error().Msg("Invalid route data: empty routes or legs")
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

	// Convert to kilometers and round to 2 decimal places
	totalDistanceKm := float64(totalDistance) / 1000
	totalDistance = int(math.Round(totalDistanceKm*100)) / 100

	log.Debug().
		Int("totalDistance", totalDistance).
		Int("totalDuration", totalDuration).
		Msg("Calculated total distance and duration")

	endTime := startTime.Add(time.Duration(totalDuration) * time.Second)

	var rideOfferID uuid.UUID
	err := r.db.Transaction(func(tx *gorm.DB) error {
		log.Debug().Msg("Starting database transaction")

		var vehicle migration.Vehicle
		if err := r.db.Preload("VehicleType").Where("id = ? AND user_id = ?", vehicleID, userID).First(&vehicle).Error; err != nil {
			log.Error().Err(err).Msg("Failed to fetch vehicle")
			return err
		}
		log.Debug().Interface("vehicle", vehicle).Msg("Fetched vehicle")

		// var existingRideOfferCount int64
		// err := tx.Model(&migration.RideOffer{}).
		// 	Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
		// 		userID, startTime, endTime, startTime, endTime, startTime, endTime).
		// 	Count(&existingRideOfferCount).Error
		// if err != nil {
		// 	log.Error().Err(err).Msg("Error checking for existing ride offers")
		// 	return fmt.Errorf("error checking for existing ride offers: %w", err)
		// }
		// if existingRideOfferCount > 0 {
		// 	log.Warn().Int64("count", existingRideOfferCount).Msg("Ride offer already exists for the user in that time frame")
		// 	return errors.New("ride offer already exists for the user in that time frame")
		// }

		var existingRideOfferCount int64
		err := tx.Model(&migration.RideOffer{}).
			Where("user_id = ? AND status NOT IN ('completed', 'cancelled') AND "+
				"((start_time < ? AND end_time > ?) OR "+
				"(start_time >= ? AND start_time < ?) OR "+
				"(end_time > ? AND end_time <= ?))",
				userID, endTime, startTime, startTime, endTime, startTime, endTime).
			Count(&existingRideOfferCount).Error
		if err != nil {
			log.Error().Err(err).Msg("Error checking for existing ride offers")
			return fmt.Errorf("error checking for existing ride offers: %w", err)
		}
		if existingRideOfferCount > 0 {
			log.Warn().Int64("count", existingRideOfferCount).Msg("Ride offer already exists for the user in that time frame")
			return errors.New("ride offer already exists for the user in that time frame")
		}

		var existingRideRequestCount int64
		err = tx.Model(&migration.RideRequest{}).
			Where("user_id = ? AND status NOT IN ('completed', 'cancelled') AND "+
				"((start_time < ? AND end_time > ?) OR "+
				"(start_time >= ? AND start_time < ?) OR "+
				"(end_time > ? AND end_time <= ?))",
				userID, endTime, startTime, startTime, endTime, startTime, endTime).
			Count(&existingRideRequestCount).Error
		if err != nil {
			log.Error().Err(err).Msg("Error checking for existing ride requests")
			return fmt.Errorf("error checking for existing ride requests: %w", err)
		}
		if existingRideRequestCount > 0 {
			log.Warn().Int64("count", existingRideRequestCount).Msg("Ride request already exists for the user in that time frame")
			return errors.New("ride request already exists for the user in that time frame")
		}

		// var existingRideRequestCount int64
		// err = tx.Model(&migration.RideRequest{}).
		// 	Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
		// 		userID, startTime, endTime, startTime, endTime, startTime, endTime).
		// 	Count(&existingRideRequestCount).Error
		// if err != nil {
		// 	log.Error().Err(err).Msg("Error checking for existing ride requests")
		// 	return fmt.Errorf("error checking for existing ride requests: %w", err)
		// }
		// if existingRideRequestCount > 0 {
		// 	log.Warn().Int64("count", existingRideRequestCount).Msg("Ride request already exists for the user in that time frame")
		// 	return errors.New("ride request already exists for the user in that time frame")
		// }

		var fuelPrice float64
		if err := tx.Model(&migration.FuelPrice{}).
			Select("price").
			Where("fuel_type = ?", "Xăng RON 95-III").
			First(&fuelPrice).Error; err != nil {
			log.Error().Err(err).Msg("Failed to fetch fuel price")
			return fmt.Errorf("failed to fetch fuel price: %w", err)
		}
		log.Debug().Float64("fuelPrice", fuelPrice).Msg("Fetched fuel price")

		// Calculate the initial fare as a float64 for precision
		fare := (vehicle.FuelConsumed / 100) * fuelPrice * float64(totalDistance)
		log.Debug().Float64("fare", fare).Msg("Calculated initial fare")

		// Round the fare to the nearest 1000 VND
		roundedFare := math.Round(fare/1000) * 1000
		log.Debug().Float64("roundedFare", roundedFare).Msg("Rounded fare to nearest 1000 VND")

		// Ensure the minimum fare is 1000 VND
		if roundedFare < 1000 {
			roundedFare = 1000
		}

		// Convert the rounded fare to int64
		realFare := int64(roundedFare)
		log.Debug().Int64("realFare", realFare).Msg("Final fare as int64")

		decodePolyline := helper.DecodePolyline(firstRoute.Overview_polyline.Points)
		startLocation := schemas.Point{
			Lat: firstLeg.Start_location.Lat,
			Lng: firstLeg.Start_location.Lng,
		}
		endLocation := schemas.Point{
			Lat: lastLeg.End_location.Lat,
			Lng: lastLeg.End_location.Lng,
		}

		newStartLocaton, newEndLocation := helper.FindClosestPoints(decodePolyline, startLocation, endLocation)
		log.Debug().
			Interface("newStartLocation", newStartLocaton).
			Interface("newEndLocation", newEndLocation).
			Msg("Found closest points on route")

		rideOffer := migration.RideOffer{
			UserID:                 userID,
			StartLatitude:          newStartLocaton.Lat,
			StartLongitude:         newStartLocaton.Lng,
			EndLatitude:            newEndLocation.Lat,
			EndLongitude:           newEndLocation.Lng,
			EncodedPolyline:        polyline.Polyline(firstRoute.Overview_polyline.Points),
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
			Fare:                   realFare,
		}

		if err := tx.Create(&rideOffer).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create ride offer")
			return fmt.Errorf("failed to create ride offer: %w", err)
		}
		log.Info().Str("rideOfferID", rideOffer.ID.String()).Msg("Created ride offer")

		rideOfferID = rideOffer.ID

		if len(firstRoute.Legs) < 2 {
			log.Debug().Msg("No need to create waypoints, only one leg")
			return nil
		}

		newWaypoints := make([]migration.Waypoint, 0, len(route.Geocoded_waypoints)-2)
		for i, leg := range firstRoute.Legs {
			if i == len(firstRoute.Legs)-1 {
				break
			}

			legEnd := leg.End_location

			waypoint := migration.Waypoint{
				RideOfferID:   rideOfferID,
				Latitude:      legEnd.Lat,
				Longitude:     legEnd.Lng,
				Address:       leg.End_address,
				WaypointOrder: i,
			}
			newWaypoints = append(newWaypoints, waypoint)
		}
		if err := tx.Create(&newWaypoints).Error; err != nil {
			log.Error().Err(err).Msg("Failed to create waypoints")
			return fmt.Errorf("failed to create waypoints: %w", err)
		}
		log.Debug().Int("waypointCount", len(newWaypoints)).Msg("Created waypoints")

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Transaction failed")
		return uuid.Nil, err
	}

	log.Info().Str("rideOfferID", rideOfferID.String()).Msg("Successfully created ride offer")
	return rideOfferID, nil
}

// func (r *MapsRepository) CreateGiveRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time, vehicleID uuid.UUID) (uuid.UUID, error) {
// 	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
// 		return uuid.Nil, errors.New("invalid route data")
// 	}

// 	firstRoute := route.Routes[0]
// 	firstLeg := firstRoute.Legs[0]
// 	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

// 	totalDistance, totalDuration := 0, 0
// 	for _, leg := range firstRoute.Legs {
// 		totalDistance += leg.Distance.Value
// 		totalDuration += leg.Duration.Value
// 	}
// 	totalDistance /= 1000 // Convert to kilometers

// 	endTime := startTime.Add(time.Duration(totalDuration) * time.Second)

// 	var rideOfferID uuid.UUID
// 	err := r.db.Transaction(func(tx *gorm.DB) error {
// 		// Check if the vehicle exists and belongs to the user
// 		// Check if the vehicle exists and belongs to the user
// 		var vehicle migration.Vehicle
// 		if err := r.db.Preload("VehicleType").Where("id = ? AND user_id = ?", vehicleID, userID).First(&vehicle).Error; err != nil {
// 			return err
// 		}

// 		// Check for overlapping ride offers
// 		var existingRideOfferCount int64
// 		err := tx.Model(&migration.RideOffer{}).
// 			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
// 				userID, startTime, endTime, startTime, endTime, startTime, endTime).
// 			Count(&existingRideOfferCount).Error
// 		if err != nil {
// 			return fmt.Errorf("error checking for existing ride offers: %w", err)
// 		}
// 		if existingRideOfferCount > 0 {
// 			return errors.New("ride offer already exists for the user in that time frame")
// 		}

// 		// Check for overlapping ride requests
// 		var existingRideRequestCount int64
// 		err = tx.Model(&migration.RideRequest{}).
// 			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
// 				userID, startTime, endTime, startTime, endTime, startTime, endTime).
// 			Count(&existingRideRequestCount).Error
// 		if err != nil {
// 			return fmt.Errorf("error checking for existing ride requests: %w", err)
// 		}
// 		if existingRideRequestCount > 0 {
// 			return errors.New("ride request already exists for the user in that time frame")
// 		}

// 		// Fetch fuel price
// 		var fuelPrice float64
// 		if err := tx.Model(&migration.FuelPrice{}).
// 			Select("price").
// 			Where("fuel_type = ?", "Xăng RON 95-III").
// 			First(&fuelPrice).Error; err != nil {
// 			return fmt.Errorf("failed to fetch fuel price: %w", err)
// 		}

// 		fare := (vehicle.FuelConsumed / 100) * fuelPrice * float64(totalDistance)
// 		// Decode the polyline and get the list of coordinates
// 		decodePolyline := helper.DecodePolyline(firstRoute.Overview_polyline.Points)
// 		startLocation := schemas.Point{
// 			Lat: firstLeg.Start_location.Lat,
// 			Lng: firstLeg.Start_location.Lng,
// 		}
// 		endLocation := schemas.Point{
// 			Lat: lastLeg.End_location.Lat,
// 			Lng: lastLeg.End_location.Lng,
// 		}

// 		// Get the correct start and end locations on the route
// 		newStartLocaton, newEndLocation := helper.FindClosestPoints(decodePolyline, startLocation, endLocation)

// 		// Create ride offer
// 		// rideOffer := migration.RideOffer{
// 		// 	UserID:                 userID,
// 		// 	StartLatitude:          firstLeg.Start_location.Lat,
// 		// 	StartLongitude:         firstLeg.Start_location.Lng,
// 		// 	EndLatitude:            lastLeg.End_location.Lat,
// 		// 	EndLongitude:           lastLeg.End_location.Lng,
// 		// 	EncodedPolyline:        polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
// 		// 	DriverCurrentLatitude:  currentLocation.Lat,
// 		// 	DriverCurrentLongitude: currentLocation.Lng,
// 		// 	StartAddress:           firstLeg.Start_address,
// 		// 	EndAddress:             lastLeg.End_address,
// 		// 	Distance:               float64(totalDistance),
// 		// 	Duration:               totalDuration,
// 		// 	Status:                 "created",
// 		// 	StartTime:              startTime,
// 		// 	EndTime:                endTime,
// 		// 	VehicleID:              vehicleID,
// 		// 	Fare:                   fare,
// 		// }

// 		rideOffer := migration.RideOffer{
// 			UserID:                 userID,
// 			StartLatitude:          newStartLocaton.Lat,
// 			StartLongitude:         newStartLocaton.Lng,
// 			EndLatitude:            newEndLocation.Lat,
// 			EndLongitude:           newEndLocation.Lng,
// 			EncodedPolyline:        polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
// 			DriverCurrentLatitude:  currentLocation.Lat,
// 			DriverCurrentLongitude: currentLocation.Lng,
// 			StartAddress:           firstLeg.Start_address,
// 			EndAddress:             lastLeg.End_address,
// 			Distance:               float64(totalDistance),
// 			Duration:               totalDuration,
// 			Status:                 "created",
// 			StartTime:              startTime,
// 			EndTime:                endTime,
// 			VehicleID:              vehicleID,
// 			Fare:                   fare,
// 		}

// 		if err := tx.Create(&rideOffer).Error; err != nil {
// 			return fmt.Errorf("failed to create ride offer: %w", err)
// 		}

// 		rideOfferID = rideOffer.ID

// 		// Only create waypoints if there are more than 2 legs
// 		if len(firstRoute.Legs) < 2 {
// 			// no need to create waypoints if there are only 1
// 			return nil
// 		}

// 		// Create waypoints
// 		newWaypoints := make([]migration.Waypoint, 0, len(route.Geocoded_waypoints)-2) // Exclude the start and end locations
// 		// TODO: Find closest points on the route for the start and end locations
// 		// only need to store waypoints that are not the start or end locations
// 		// And the waypoints is end location of the previous leg and start location of the next leg
// 		for i, leg := range firstRoute.Legs {
// 			// Store the end location of the leg as a waypoint
// 			// only not store the last leg end location
// 			if i == len(firstRoute.Legs)-1 {
// 				break
// 			}

// 			legEnd := leg.End_location

// 			waypoint := migration.Waypoint{
// 				RideOfferID:   rideOfferID,
// 				Latitude:      legEnd.Lat,
// 				Longitude:     legEnd.Lng,
// 				Address:       leg.End_address, // Store the address of the end location of the leg
// 				WaypointOrder: i,
// 			}
// 			newWaypoints = append(newWaypoints, waypoint)
// 		}
// 		if err := tx.Create(&newWaypoints).Error; err != nil {
// 			return fmt.Errorf("failed to create waypoints: %w", err)
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		return uuid.Nil, err
// 	}

// 	return rideOfferID, nil
// }

func (r *MapsRepository) CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error) {
	log.Debug().
		Interface("route", route).
		Str("userID", userID.String()).
		Interface("currentLocation", currentLocation).
		Time("startTime", startTime).
		Msg("CreateHitchRide function called")

	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
		log.Error().Msg("Invalid route data: empty routes or legs")
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
	// Convert to kilometers and round to 2 decimal places
	totalDistanceKm := float64(totalDistance) / 1000
	totalDistance = int(math.Round(totalDistanceKm*100)) / 100

	endTime := startTime.Add(time.Duration(totalDuration) * time.Second)

	log.Debug().
		Int("totalDistance", totalDistance).
		Int("totalDuration", totalDuration).
		Time("endTime", endTime).
		Msg("Calculated total distance, duration, and end time")

	var rideRequestID uuid.UUID
	err := r.db.Transaction(func(tx *gorm.DB) error {
		log.Debug().Msg("Starting database transaction")

		// var existingRideRequestCount int64
		// err := tx.Model(&migration.RideRequest{}).
		// 	Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
		// 		userID, startTime, endTime, startTime, endTime, startTime, endTime).
		// 	Count(&existingRideRequestCount).Error
		// if err != nil {
		// 	log.Error().Err(err).Msg("Error checking for existing ride requests")
		// 	return fmt.Errorf("error checking for existing ride requests: %w", err)
		// }
		// if existingRideRequestCount > 0 {
		// 	log.Warn().Int64("count", existingRideRequestCount).Msg("Ride request already exists for the user in that time frame")
		// 	return errors.New("ride request already exists for the user in that time frame")
		// }

		var existingRideRequestCount int64
		err := tx.Model(&migration.RideRequest{}).
			Where("user_id = ? AND status NOT IN ('completed', 'cancelled') AND "+
				"((start_time < ? AND end_time > ?) OR "+
				"(start_time >= ? AND start_time < ?) OR "+
				"(end_time > ? AND end_time <= ?))",
				userID, endTime, startTime, startTime, endTime, startTime, endTime).
			Count(&existingRideRequestCount).Error
		if err != nil {
			log.Error().Err(err).Msg("Error checking for existing ride requests")
			return fmt.Errorf("error checking for existing ride requests: %w", err)
		}
		if existingRideRequestCount > 0 {
			log.Warn().Int64("count", existingRideRequestCount).Msg("Ride request already exists for the user in that time frame")
			return errors.New("ride request already exists for the user in that time frame")
		}

		// var existingRideOfferCount int64
		// err = tx.Model(&migration.RideOffer{}).
		// 	Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
		// 		userID, startTime, endTime, startTime, endTime, startTime, endTime).
		// 	Count(&existingRideOfferCount).Error
		// if err != nil {
		// 	log.Error().Err(err).Msg("Error checking for existing ride offers")
		// 	return fmt.Errorf("error checking for existing ride offers: %w", err)
		// }
		// if existingRideOfferCount > 0 {
		// 	log.Warn().Int64("count", existingRideOfferCount).Msg("Ride offer already exists for the user in that time frame")
		// 	return errors.New("ride offer already exists for the user in that time frame")
		// }

		var existingRideOfferCount int64
		err = tx.Model(&migration.RideOffer{}).
			Where("user_id = ? AND status NOT IN ('completed', 'cancelled') AND "+
				"((start_time < ? AND end_time > ?) OR "+
				"(start_time >= ? AND start_time < ?) OR "+
				"(end_time > ? AND end_time <= ?))",
				userID, endTime, startTime, startTime, endTime, startTime, endTime).
			Count(&existingRideOfferCount).Error
		if err != nil {
			log.Error().Err(err).Msg("Error checking for existing ride offers")
			return fmt.Errorf("error checking for existing ride offers: %w", err)
		}
		if existingRideOfferCount > 0 {
			log.Warn().Int64("count", existingRideOfferCount).Msg("Ride offer already exists for the user in that time frame")
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

		newStartLocaton, newEndLocation := helper.FindClosestPoints(decodePolyline, startLocation, endLocation)
		log.Debug().
			Interface("newStartLocation", newStartLocaton).
			Interface("newEndLocation", newEndLocation).
			Msg("Found closest points on route")

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
			EncodedPolyline:       polyline.Polyline(firstRoute.Overview_polyline.Points),
			Distance:              float64(totalDistance),
			Duration:              totalDuration,
			StartTime:             startTime,
			EndTime:               endTime,
		}

		if err := tx.Create(&rideRequest).Error; err != nil {
			log.Error().Err(err).Msg("Error creating ride request")
			return fmt.Errorf("error creating ride request: %w", err)
		}
		rideRequestID = rideRequest.ID
		log.Info().Str("rideRequestID", rideRequestID.String()).Msg("Created ride request")

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Transaction failed")
		return uuid.Nil, err
	}

	log.Info().Str("rideRequestID", rideRequestID.String()).Msg("Successfully created hitch ride request")
	return rideRequestID, nil
}

// func (r *MapsRepository) CreateHitchRide(route schemas.GoongDirectionsResponse, userID uuid.UUID, currentLocation schemas.Point, startTime time.Time) (uuid.UUID, error) {
// 	if len(route.Routes) == 0 || len(route.Routes[0].Legs) == 0 {
// 		return uuid.Nil, errors.New("invalid route data")
// 	}

// 	firstRoute := route.Routes[0]
// 	firstLeg := firstRoute.Legs[0]
// 	lastLeg := firstRoute.Legs[len(firstRoute.Legs)-1]

// 	totalDistance, totalDuration := 0, 0
// 	for _, leg := range firstRoute.Legs {
// 		totalDistance += leg.Distance.Value
// 		totalDuration += leg.Duration.Value
// 	}
// 	totalDistance /= 1000 // Convert to kilometers
// 	endTime := startTime.Add(time.Duration(totalDuration) * time.Second)

// 	var rideRequestID uuid.UUID
// 	err := r.db.Transaction(func(tx *gorm.DB) error {
// 		// Check if any ride request exists for the user that overlaps with the new time frame
// 		var existingRideRequestCount int64
// 		err := tx.Model(&migration.RideRequest{}).
// 			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
// 				userID, startTime, endTime, startTime, endTime, startTime, endTime).
// 			Count(&existingRideRequestCount).Error
// 		if err != nil {
// 			return fmt.Errorf("error checking for existing ride requests: %w", err)
// 		}
// 		if existingRideRequestCount > 0 {
// 			return errors.New("ride request already exists for the user in that time frame")
// 		}

// 		// Check if any ride offer exists for the user that overlaps with the new time frame
// 		var existingRideOfferCount int64
// 		err = tx.Model(&migration.RideOffer{}).
// 			Where("user_id = ? AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?) OR (start_time <= ? AND end_time >= ?))",
// 				userID, startTime, endTime, startTime, endTime, startTime, endTime).
// 			Count(&existingRideOfferCount).Error
// 		if err != nil {
// 			return fmt.Errorf("error checking for existing ride offers: %w", err)
// 		}
// 		if existingRideOfferCount > 0 {
// 			return errors.New("ride offer already exists for the user in that time frame")
// 		}

// 		decodePolyline := helper.DecodePolyline(firstRoute.Overview_polyline.Points)
// 		startLocation := schemas.Point{
// 			Lat: firstLeg.Start_location.Lat,
// 			Lng: firstLeg.Start_location.Lng,
// 		}
// 		endLocation := schemas.Point{
// 			Lat: lastLeg.End_location.Lat,
// 			Lng: lastLeg.End_location.Lng,
// 		}

// 		// Get the correct start and end locations on the route
// 		newStartLocaton, newEndLocation := helper.FindClosestPoints(decodePolyline, startLocation, endLocation)

// 		// Create ride request
// 		// rideRequest := migration.RideRequest{
// 		// 	UserID:                userID,
// 		// 	StartLatitude:         firstLeg.Start_location.Lat,
// 		// 	StartLongitude:        firstLeg.Start_location.Lng,
// 		// 	EndLatitude:           lastLeg.End_location.Lat,
// 		// 	EndLongitude:          lastLeg.End_location.Lng,
// 		// 	RiderCurrentLatitude:  currentLocation.Lat,
// 		// 	RiderCurrentLongitude: currentLocation.Lng,
// 		// 	StartAddress:          firstLeg.Start_address,
// 		// 	EndAddress:            lastLeg.End_address,
// 		// 	Status:                "created",
// 		// 	EncodedPolyline:       polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
// 		// 	Distance:              float64(totalDistance),
// 		// 	Duration:              totalDuration,
// 		// 	StartTime:             startTime,
// 		// 	EndTime:               endTime,
// 		// }

// 		rideRequest := migration.RideRequest{
// 			UserID:                userID,
// 			StartLatitude:         newStartLocaton.Lat,
// 			StartLongitude:        newStartLocaton.Lng,
// 			EndLatitude:           newEndLocation.Lat,
// 			EndLongitude:          newEndLocation.Lng,
// 			RiderCurrentLatitude:  currentLocation.Lat,
// 			RiderCurrentLongitude: currentLocation.Lng,
// 			StartAddress:          firstLeg.Start_address,
// 			EndAddress:            lastLeg.End_address,
// 			Status:                "created",
// 			EncodedPolyline:       polyline.Polyline(firstRoute.Overview_polyline.Points), // Use gorm.Expr to prevent escaping
// 			Distance:              float64(totalDistance),
// 			Duration:              totalDuration,
// 			StartTime:             startTime,
// 			EndTime:               endTime,
// 		}

// 		if err := tx.Create(&rideRequest).Error; err != nil {
// 			return fmt.Errorf("error creating ride request: %w", err)
// 		}
// 		rideRequestID = rideRequest.ID
// 		return nil
// 	})

// 	if err != nil {
// 		return uuid.Nil, err
// 	}
// 	return rideRequestID, nil
// }

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
