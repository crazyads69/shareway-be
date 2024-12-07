package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"shareway/helper"
	"shareway/infra/db/migration"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	MaxRetry = 5 // Maximum number of retries for fetching data from Goong API
)

type IMapService interface {
	GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool, currentLocation string) (schemas.GoongAutoCompleteResponse, error)
	CreateGiveRide(ctx context.Context, input schemas.GiveRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error)
	CreateHitchRide(ctx context.Context, input schemas.HitchRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error)
	GetGeoCode(ctx context.Context, point schemas.Point, currentLocation schemas.Point) (schemas.GeoCodeLocationResponse, error)
	GetLocationFromPlaceID(ctx context.Context, placeID string) (schemas.Point, error)
	GetRideOfferDetails(ctx context.Context, rideOfferID uuid.UUID) (migration.RideOffer, error)
	GetRideRequestDetails(ctx context.Context, rideRequestID uuid.UUID) (migration.RideRequest, error)
	GetDistanceFromCurrentLocation(ctx context.Context, currentLocation schemas.Point, destinationPoint []schemas.Point) (schemas.GoongDistanceMatrixResponse, error)
	SuggestRideRequests(ctx context.Context, userID uuid.UUID, rideOfferID uuid.UUID) ([]migration.RideRequest, error)
	SuggestRideOffers(ctx context.Context, userID uuid.UUID, rideRequestID uuid.UUID) ([]migration.RideOffer, error)
	GetAllWaypoints(rideOfferID uuid.UUID) ([]migration.Waypoint, error)
}

type MapService struct {
	repo        repository.IMapsRepository
	cfg         util.Config
	redisClient *redis.Client
}

func NewMapService(repo repository.IMapsRepository, cfg util.Config, redisClient *redis.Client) IMapService {
	return &MapService{
		repo:        repo,
		cfg:         cfg,
		redisClient: redisClient,
	}
}

// GetLocationFromPlaceID returns the location (latitude, longitude) of the given place ID
func (s *MapService) GetLocationFromPlaceID(ctx context.Context, placeID string) (schemas.Point, error) {
	baseURL, err := url.Parse(fmt.Sprintf("%s/place/detail", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.Point{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key":  {s.cfg.GoongAPIKey},
		"place_id": {placeID},
	}
	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	var response schemas.GoongPlaceDetailResponse
	maxRetries := MaxRetry
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.Point{}, err
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return schemas.Point{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.Point{}, err
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.Point{}, err
		}

		return schemas.Point{
			Lat: response.Result.Geometry.Location.Lat,
			Lng: response.Result.Geometry.Location.Lng,
		}, nil
	}

	// If max retries reached and still no success, return an error
	return schemas.Point{}, fmt.Errorf("max retries reached, unable to get location data")
}

// GetAutoComplete returns the auto-complete results for the given input
func (s *MapService) GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool, currentLocation string) (schemas.GoongAutoCompleteResponse, error) {
	// Build the request URL
	baseURL, err := url.Parse(fmt.Sprintf("%s/place/autocomplete", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"input":   {input},
	}

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	} else {
		params.Set("limit", strconv.Itoa(4))
	}
	if location != "" {
		params.Set("location", location)
	}
	if radius > 0 {
		params.Set("radius", strconv.Itoa(radius))
	} else {
		params.Set("radius", strconv.Itoa(50))
	}
	if moreCompound {
		params.Set("more_compound", "true")
	}

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	var response schemas.GoongAutoCompleteResponse
	maxRetries := MaxRetry
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("failed to fetch from Goong API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("unexpected status code from Goong API: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// If we've reached here, we've successfully got and parsed the response
		break
	}

	if currentLocation != "" {
		currentLocationPoint := helper.ConvertStringToLocation(currentLocation)

		destinationPoints := make([]schemas.Point, len(response.Predictions))
		for i, prediction := range response.Predictions {
			point, err := s.GetLocationFromPlaceID(ctx, prediction.PlaceID)
			if err != nil {
				log.Printf("Failed to get location for place ID %s: %v", prediction.PlaceID, err)
				continue
			}
			destinationPoints[i] = point
		}

		distanceMatrix, err := s.GetDistanceFromCurrentLocation(ctx, currentLocationPoint, destinationPoints)
		if err != nil {
			log.Printf("Failed to get distance matrix: %v", err)
		} else {
			for i := range response.Predictions {
				if i < len(distanceMatrix.Rows[0].Elements) {
					// Convert to km and round to 2 decimal places
					distanceKm := float64(distanceMatrix.Rows[0].Elements[i].Distance.Value) / 1000
					roundedDistance := math.Round(distanceKm*100) / 100
					response.Predictions[i].Distance = roundedDistance
				}
			}
		}
	}

	return response, nil
}

// CreateGiveRide creates a ride offer based on the given input
func (s *MapService) CreateGiveRide(ctx context.Context, input schemas.GiveRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error) {
	points := make([]schemas.Point, len(input.PlaceList))
	for i, placeID := range input.PlaceList {
		point, err := s.GetLocationFromPlaceID(ctx, placeID)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to get location for place ID %s: %w", placeID, err)
		}
		points[i] = point
	}

	baseURL, err := url.Parse(fmt.Sprintf("%s/direction", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"origin":  {fmt.Sprintf("%f,%f", points[0].Lat, points[0].Lng)},
		"vehicle": {"hd"},
	}

	destinations := make([]string, len(points)-1)
	for i := 1; i < len(points); i++ {
		destinations[i-1] = fmt.Sprintf("%f,%f", points[i].Lat, points[i].Lng)
	}
	params.Set("destination", strings.Join(destinations, ";"))

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	var response schemas.GoongDirectionsResponse
	maxRetries := MaxRetry
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to fetch from Goong API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// If we've reached here, we've successfully got and parsed the response
		break
	}

	currentLocation := schemas.Point{
		Lat: points[0].Lat,
		Lng: points[0].Lng,
	}

	// Check start_time from input and set the ride request status accordingly
	// If start_time is not provided, the ride is immediate
	var startTime time.Time
	if input.StartTime != "" {
		// Parse the start time as GMT+7
		location, err := time.LoadLocation("Asia/Bangkok") // GMT+7
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to load location: %w", err)
		}

		// Parse the time in the GMT+7 location
		startTime, err = time.ParseInLocation("2006-01-02T15:04:05.999999", input.StartTime, location)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to parse start time: %w", err)
		}

		// Convert to UTC
		startTime = startTime.UTC()
	} else {
		startTime = time.Now().UTC()
	}

	rideOfferID, err := s.repo.CreateGiveRide(response, userID, currentLocation, startTime, input.VehicleID)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	return response, rideOfferID, nil
}

// CreateHitchRide creates a hitch ride request based on the given input
func (s *MapService) CreateHitchRide(ctx context.Context, input schemas.HitchRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error) {
	points := make([]schemas.Point, len(input.PlaceList))
	for i, placeID := range input.PlaceList {
		point, err := s.GetLocationFromPlaceID(ctx, placeID)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to get location for place ID %s: %w", placeID, err)
		}
		points[i] = point
	}

	baseURL, err := url.Parse(fmt.Sprintf("%s/direction", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"origin":  {fmt.Sprintf("%f,%f", points[0].Lat, points[0].Lng)},
		"vehicle": {"hd"},
	}

	destinations := make([]string, len(points)-1)
	for i := 1; i < len(points); i++ {
		destinations[i-1] = fmt.Sprintf("%f,%f", points[i].Lat, points[i].Lng)
	}
	params.Set("destination", strings.Join(destinations, ";"))

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	var response schemas.GoongDirectionsResponse
	maxRetries := MaxRetry
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to fetch from Goong API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// If we've reached here, we've successfully got and parsed the response
		break
	}

	currentLocation := schemas.Point{
		Lat: points[0].Lat,
		Lng: points[0].Lng,
	}

	// Check start_time from input and set the ride request status accordingly
	// If start_time is not provided, the ride is immediate
	// Check start_time from input and set the ride request status accordingly
	// If start_time is not provided, the ride is immediate
	var startTime time.Time
	if input.StartTime != "" {
		// Parse the start time as GMT+7
		location, err := time.LoadLocation("Asia/Bangkok") // GMT+7
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to load location: %w", err)
		}

		// Parse the time in the GMT+7 location
		startTime, err = time.ParseInLocation("2006-01-02T15:04:05.999999", input.StartTime, location)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to parse start time: %w", err)
		}

		// Convert to UTC
		startTime = startTime.UTC()
	} else {
		startTime = time.Now().UTC()
	}

	rideRequestID, err := s.repo.CreateHitchRide(response, userID, currentLocation, startTime)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	return response, rideRequestID, nil
}

// GetGeoCode returns the geocode information for the given point
func (s *MapService) GetGeoCode(ctx context.Context, point schemas.Point, currentLocation schemas.Point) (schemas.GeoCodeLocationResponse, error) {
	baseURL, err := url.Parse(fmt.Sprintf("%s/geocode", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GeoCodeLocationResponse{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"latlng":  {fmt.Sprintf("%f,%f", point.Lat, point.Lng)},
	}
	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	var response schemas.GoongReverseGeocodeResponse
	maxRetries := MaxRetry
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("http get error: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("read body error: %w", err)
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("unmarshal error: %w", err)
		}

		// If we've reached here, we've successfully got and parsed the response
		break
	}

	optimizedResults := schemas.GeoCodeLocationResponse{
		Results: make([]schemas.GeoCodeLocation, len(response.Results)),
	}
	destinationPoints := make([]schemas.Point, len(response.Results))

	for i, result := range response.Results {
		addressParts := strings.SplitN(result.FormattedAddress, ",", 2)
		optimizedResults.Results[i] = schemas.GeoCodeLocation{
			PlaceID:          result.PlaceID,
			FormattedAddress: result.FormattedAddress,
			Latitude:         result.Geometry.Location.Lat,
			Longitude:        result.Geometry.Location.Lng,
			MainAddress:      strings.TrimSpace(addressParts[0]),
			SecondaryAddress: strings.TrimSpace(strings.Join(addressParts[1:], ",")),
		}
		destinationPoints[i] = schemas.Point{
			Lat: result.Geometry.Location.Lat,
			Lng: result.Geometry.Location.Lng,
		}
	}

	// Calculate the distance from the current location
	distanceMatrix, err := s.GetDistanceFromCurrentLocation(ctx, currentLocation, destinationPoints)
	if err != nil {
		return schemas.GeoCodeLocationResponse{}, err
	}

	for i := range optimizedResults.Results {
		// Convert to km and round to 2 decimal places
		distanceKm := float64(distanceMatrix.Rows[0].Elements[i].Distance.Value) / 1000
		roundedDistance := math.Round(distanceKm*100) / 100
		optimizedResults.Results[i].Distance = roundedDistance
	}

	return optimizedResults, nil
}

// GetDistanceFromCurrentLocation returns the distance matrix from the current location to the destination points
func (s *MapService) GetDistanceFromCurrentLocation(ctx context.Context, currentLocation schemas.Point, destinationPoints []schemas.Point) (schemas.GoongDistanceMatrixResponse, error) {
	baseURL, err := url.Parse(fmt.Sprintf("%s/distancematrix", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"origins": {fmt.Sprintf("%f,%f", currentLocation.Lat, currentLocation.Lng)},
		"vehicle": {"hd"}, // for hail driving vehicle
	}

	destinations := make([]string, len(destinationPoints))
	for i, point := range destinationPoints {
		destinations[i] = fmt.Sprintf("%f,%f", point.Lat, point.Lng)
	}
	params.Set("destinations", strings.Join(destinations, "|"))

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	var response schemas.GoongDistanceMatrixResponse
	maxRetries := MaxRetry
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("http get error: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("read body error: %w", err)
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("unmarshal error: %w", err)
		}

		// If we've reached here, we've successfully got and parsed the response
		return response, nil
	}

	// If we've exhausted all retries
	return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("max retries reached, unable to get distance matrix")
}

// GetRideOfferDetails returns the ride offer details for the given ride offer ID
func (s *MapService) GetRideOfferDetails(ctx context.Context, rideOfferID uuid.UUID) (migration.RideOffer, error) {
	return s.repo.GetRideOfferDetails(rideOfferID)
}

// GetRideRequestDetails returns the ride request details for the given ride request ID
func (s *MapService) GetRideRequestDetails(ctx context.Context, rideRequestID uuid.UUID) (migration.RideRequest, error) {
	return s.repo.GetRideRequestDetails(rideRequestID)
}

// SuggestRideRequests returns the suggested ride requests for the given user and ride offer
func (s *MapService) SuggestRideRequests(ctx context.Context, userID uuid.UUID, rideOfferID uuid.UUID) ([]migration.RideRequest, error) {
	return s.repo.SuggestRideRequests(userID, rideOfferID)
}

// SuggestRideOffers returns the suggested ride offers for the given user and ride request
func (s *MapService) SuggestRideOffers(ctx context.Context, userID uuid.UUID, rideRequestID uuid.UUID) ([]migration.RideOffer, error) {
	return s.repo.SuggestRideOffers(userID, rideRequestID)
}

// GetAllWaypoints returns all waypoints for the given ride offer ID
func (s *MapService) GetAllWaypoints(rideOfferID uuid.UUID) ([]migration.Waypoint, error) {
	return s.repo.GetAllWaypoints(rideOfferID)
}

// Make sure MapsService implements IMapsService
var _ IMapService = (*MapService)(nil)
