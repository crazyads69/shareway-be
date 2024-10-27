package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"shareway/helper"
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
	GetGeoCode(ctx context.Context, point schemas.Point) (schemas.GoongReverseGeocodeResponse, error)
	GetLocationFromPlaceID(ctx context.Context, placeID string) (schemas.Point, error)
	GetDistanceFromCurrentLocation(ctx context.Context, currentLocation schemas.Point, destinationPoint []schemas.Point) (schemas.GoongDistanceMatrixResponse, error)
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
	}
	if location != "" {
		params.Set("location", location)
	}
	if radius > 0 {
		params.Set("radius", strconv.Itoa(radius))
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
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}
	} else {
		// If cache miss, fetch data from Goong API
		resp, err := http.Get(url)
		if err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}

		// Cache the response
		if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheAutocompleteDuration)).Err(); err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}
	}

	if currentLocation != "" {
		currentLocationPoint := helper.ConvertStringToLocation(currentLocation)
		if err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}

		destinationPoints := make([]schemas.Point, len(response.Predictions))
		for i, prediction := range response.Predictions {
			point, err := s.GetLocationFromPlaceID(ctx, prediction.PlaceID)
			if err != nil {
				return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("failed to get location for place ID %s: %w", prediction.PlaceID, err)
			}
			destinationPoints[i] = point
		}

		distanceMatrix, err := s.GetDistanceFromCurrentLocation(ctx, currentLocationPoint, destinationPoints)
		if err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}

		for i := range response.Predictions {
			// Explicitly convert int to float64
			response.Predictions[i].Distance = float64(distanceMatrix.Rows[0].Elements[i].Distance.Value)
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

	optimizedPoints := helper.OptimizeRoutePoints(points)

	baseURL, err := url.Parse(fmt.Sprintf("%s/direction", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"origin":  {fmt.Sprintf("%f,%f", optimizedPoints[0].Lat, optimizedPoints[0].Lng)},
		"vehicle": {"hd"},
	}

	destinations := make([]string, len(optimizedPoints)-1)
	for i := 1; i < len(optimizedPoints); i++ {
		destinations[i-1] = fmt.Sprintf("%f,%f", optimizedPoints[i].Lat, optimizedPoints[i].Lng)
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
		Lat: optimizedPoints[0].Lat,
		Lng: optimizedPoints[0].Lng,
	}

	rideOfferID, err := s.repo.CreateGiveRide(response, userID, currentLocation)
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

	optimizedPoints := helper.OptimizeRoutePoints(points)

	baseURL, err := url.Parse(fmt.Sprintf("%s/direction", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"api_key": {s.cfg.GoongAPIKey},
		"origin":  {fmt.Sprintf("%f,%f", optimizedPoints[0].Lat, optimizedPoints[0].Lng)},
		"vehicle": {"hd"},
	}

	destinations := make([]string, len(optimizedPoints)-1)
	for i := 1; i < len(optimizedPoints); i++ {
		destinations[i-1] = fmt.Sprintf("%f,%f", optimizedPoints[i].Lat, optimizedPoints[i].Lng)
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
		Lat: optimizedPoints[0].Lat,
		Lng: optimizedPoints[0].Lng,
	}

	rideRequestID, err := s.repo.CreateHitchRide(response, userID, currentLocation)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	return response, rideRequestID, nil
}

// GetGeoCode returns the geocode information for the given point
func (s *MapService) GetGeoCode(ctx context.Context, point schemas.Point) (schemas.GoongReverseGeocodeResponse, error) {
	baseURL, err := url.Parse(fmt.Sprintf("%s/geocode", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, fmt.Errorf("invalid base URL: %w", err)
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
	if err == nil {
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongReverseGeocodeResponse{}, err
		}
		return response, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.GoongReverseGeocodeResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	if err := s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCachePlaceDetailDuration)).Err(); err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	return response, nil
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
	params.Set("destinations", strings.Join(destinations, ";"))

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

// Make sure MapsService implements IMapsService
var _ IMapService = (*MapService)(nil)