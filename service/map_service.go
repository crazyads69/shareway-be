package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

	cacheKey := "maps:placeid:" + url
	var response schemas.GoongPlaceDetailResponse

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, &response); err == nil {
			return schemas.Point{
				Lat: response.Result.Geometry.Location.Lat,
				Lng: response.Result.Geometry.Location.Lng,
			}, nil
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return schemas.Point{}, err
	}
	defer resp.Body.Close()

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

	s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCachePlaceDetailDuration))

	return schemas.Point{
		Lat: response.Result.Geometry.Location.Lat,
		Lng: response.Result.Geometry.Location.Lng,
	}, nil
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

	// Check the cache
	cacheKey := fmt.Sprintf("maps:autocomplete:%s", url)
	var response schemas.GoongAutoCompleteResponse
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, &response); err == nil {
			return response, nil
		}
	}

	// If cache miss or unmarshal failed, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("failed to fetch from Goong API: %w", err)
	}
	defer resp.Body.Close()

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

	// Cache the response
	if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheAutocompleteDuration)).Err(); err != nil {
		// Log the error but don't return it
		log.Printf("Failed to cache response: %v", err)
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
					response.Predictions[i].Distance = float64(distanceMatrix.Rows[0].Elements[i].Distance.Value) / 1000 // Convert to km
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

	// optimizedPoints := helper.OptimizeRoutePoints(points)

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

	cacheKey := fmt.Sprintf("maps:directions:%s", url)
	var response schemas.GoongDirectionsResponse

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}

		if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheRouteDuration)).Err(); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
	}

	currentLocation := schemas.Point{
		Lat: points[0].Lat,
		Lng: points[0].Lng,
	}

	// Check start_time from input and set the ride request status accordingly
	// If start_time is not provided, the ride is immediate
	var startTime time.Time
	if input.StartTime != "" {
		// Parse the start time to UTC time
		startTime, err = time.Parse("2006-01-02T15:04:05.999999", input.StartTime)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to parse start time: %w", err)
		}
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

	// optimizedPoints := helper.OptimizeRoutePoints(points)

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

	cacheKey := fmt.Sprintf("maps:directions:%s", url)
	var response schemas.GoongDirectionsResponse

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}

		if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheRouteDuration)).Err(); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
	}

	currentLocation := schemas.Point{
		Lat: points[0].Lat,
		Lng: points[0].Lng,
	}

	// Check start_time from input and set the ride request status accordingly
	// If start_time is not provided, the ride is immediate
	var startTime time.Time
	if input.StartTime != "" {
		// Parse the start time to UTC time
		startTime, err = time.Parse("2006-01-02T15:04:05.999999", input.StartTime)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to parse start time: %w", err)
		}
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

	cacheKey := fmt.Sprintf("maps:geocode:%s", url)
	var response schemas.GoongReverseGeocodeResponse

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil && err != redis.Nil {
		return schemas.GeoCodeLocationResponse{}, fmt.Errorf("redis get error: %w", err)
	}

	if err == redis.Nil {
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("http get error: %w", err)
		}
		defer resp.Body.Close()

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

		if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCachePlaceDetailDuration)).Err(); err != nil {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("redis set error: %w", err)
		}
	} else {
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GeoCodeLocationResponse{}, fmt.Errorf("unmarshal cached data error: %w", err)
		}
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

	// Calculate the distance from the current
	distanceMatrix, err := s.GetDistanceFromCurrentLocation(ctx, currentLocation, destinationPoints)
	if err != nil {
		return schemas.GeoCodeLocationResponse{}, err
	}

	for i := range optimizedResults.Results {
		optimizedResults.Results[i].Distance = float64(distanceMatrix.Rows[0].Elements[i].Distance.Value) / 1000 // Convert to km
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

	cacheKey := fmt.Sprintf("maps:distancematrix:%s", url)
	var response schemas.GoongDistanceMatrixResponse

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongDistanceMatrixResponse{}, err
		}
		return response, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongDistanceMatrixResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.GoongDistanceMatrixResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.GoongDistanceMatrixResponse{}, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongDistanceMatrixResponse{}, err
	}

	if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheRouteDuration)).Err(); err != nil {
		return schemas.GoongDistanceMatrixResponse{}, err
	}

	return response, nil
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

// // SuggestRideOffers returns the suggested ride offers for the given user and ride request
func (s *MapService) SuggestRideOffers(ctx context.Context, userID uuid.UUID, rideRequestID uuid.UUID) ([]migration.RideOffer, error) {
	return s.repo.SuggestRideOffers(userID, rideRequestID)
}

// Make sure MapsService implements IMapsService
var _ IMapService = (*MapService)(nil)
