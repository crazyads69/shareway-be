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
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IMapService interface {
	GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool) (schemas.GoongAutoCompleteResponse, error)
	CreateGiveRide(ctx context.Context, input schemas.GiveRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error)
	CreateHitchRide(ctx context.Context, input schemas.HitchRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error)
	GetGeoCode(ctx context.Context, point schemas.Point) (schemas.GoongReverseGeocodeResponse, error)
	GetLocationFromPlaceID(ctx context.Context, placeID string) (schemas.Point, error)
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

func (s *MapService) GetLocationFromPlaceID(ctx context.Context, placeID string) (schemas.Point, error) {
	// Build the request URL
	baseURL, err := url.Parse(fmt.Sprintf("%s/place/detail", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.Point{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("api_key", s.cfg.GoongAPIKey)
	params.Set("place_id", placeID)

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	// Check the cache
	cacheKey := fmt.Sprintf("maps:placeid:%s", url)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var response schemas.GoongPlaceDetailResponse
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.Point{}, err
		}
		location := schemas.Point{
			Lat: response.Result.Geometry.Location.Lat,
			Lng: response.Result.Geometry.Location.Lng,
		}
		return location, nil
	}

	// If cache miss, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return schemas.Point{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.Point{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.Point{}, err
	}

	// Cache the response
	err = s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCachePlaceDetailDuration)).Err()
	if err != nil {
		return schemas.Point{}, err
	}

	var response schemas.GoongPlaceDetailResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.Point{}, err
	}

	location := schemas.Point{
		Lat: response.Result.Geometry.Location.Lat,
		Lng: response.Result.Geometry.Location.Lng,
	}
	return location, nil
}

func (s *MapService) GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool) (schemas.GoongAutoCompleteResponse, error) {
	// Build the request URL
	baseURL, err := url.Parse(fmt.Sprintf("%s/place/autocomplete", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("api_key", s.cfg.GoongAPIKey)
	params.Set("input", input)

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if location != "" {
		params.Set("location", location)
	}
	if radius > 0 {
		params.Set("radius", fmt.Sprintf("%d", radius))
	}
	if moreCompound {
		params.Set("more_compound", "true")
	}

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	// Check the cache
	cacheKey := fmt.Sprintf("maps:autocomplete:%s", url)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var response schemas.GoongAutoCompleteResponse
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongAutoCompleteResponse{}, err
		}
		return response, nil
	}

	// If cache miss, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.GoongAutoCompleteResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, err
	}

	// Cache the response
	err = s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheAutocompleteDuration)).Err()

	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, err
	}

	var response schemas.GoongAutoCompleteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongAutoCompleteResponse{}, err
	}

	return response, nil
}

func (s *MapService) CreateGiveRide(ctx context.Context, input schemas.GiveRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error) {
	var points []schemas.Point
	for _, placeID := range input.PlaceList {
		point, err := s.GetLocationFromPlaceID(ctx, placeID)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to get location for place ID %s: %w", placeID, err)
		}
		points = append(points, point)
	}

	// Build the request URL
	baseURL, err := url.Parse(fmt.Sprintf("%s/direction", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("api_key", s.cfg.GoongAPIKey)

	// Optimize the list of points for the route
	optimizedPoints := helper.OptimizeRoutePoints(points)

	// Build the params for the request
	params.Set("origin", fmt.Sprintf("%f,%f", optimizedPoints[0].Lat, optimizedPoints[0].Lng))
	// The destination is the list from the second point to the last point and separated by ';"
	var destinations string
	for i := 1; i < len(optimizedPoints); i++ {
		if i > 1 {
			destinations += ";"
		}
		destinations += fmt.Sprintf("%f,%f", optimizedPoints[i].Lat, optimizedPoints[i].Lng)
	}
	params.Set("destination", destinations)
	params.Set("vehicle", "bike")

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	// Check the cache
	cacheKey := fmt.Sprintf("maps:directions:%s", url)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var response schemas.GoongDirectionsResponse
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
		}
		// Store the create rideoffer in the database
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

	// If cache miss, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	// Cache the response
	err = s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheRouteDuration)).Err()
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	var response schemas.GoongDirectionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	currentLocation := schemas.Point{
		Lat: optimizedPoints[0].Lat,
		Lng: optimizedPoints[0].Lng,
	}
	// Store the create rideoffer in the database
	rideOfferID, err := s.repo.CreateGiveRide(response, userID, currentLocation)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	return response, rideOfferID, nil
}

func (s *MapService) CreateHitchRide(ctx context.Context, input schemas.HitchRideRequest, userID uuid.UUID) (schemas.GoongDirectionsResponse, uuid.UUID, error) {
	var points []schemas.Point
	for _, placeID := range input.PlaceList {
		point, err := s.GetLocationFromPlaceID(ctx, placeID)
		if err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("failed to get location for place ID %s: %w", placeID, err)
		}
		points = append(points, point)
	}

	// Build the request URL
	baseURL, err := url.Parse(fmt.Sprintf("%s/direction", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("api_key", s.cfg.GoongAPIKey)

	// Optimize the list of points for the route
	optimizedPoints := helper.OptimizeRoutePoints(points)

	// Build the params for the request
	params.Set("origin", fmt.Sprintf("%f,%f", optimizedPoints[0].Lat, optimizedPoints[0].Lng))
	// The destination is the list from the second point to the last point and separated by ';"
	var destinations string
	for i := 1; i < len(optimizedPoints); i++ {
		if i == len(optimizedPoints)-1 {
			destinations += fmt.Sprintf("%f,%f;", optimizedPoints[i].Lat, optimizedPoints[i].Lng)
		} else {
			destinations += fmt.Sprintf("%f,%f", optimizedPoints[i].Lat, optimizedPoints[i].Lng)
		}
	}
	params.Set("destination", destinations)
	params.Set("vehicle", "bike")

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	// Check the cache
	cacheKey := fmt.Sprintf("maps:directions:%s", url)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var response schemas.GoongDirectionsResponse
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongDirectionsResponse{}, uuid.Nil, err
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

	// If cache miss, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	// Cache the response
	err = s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheRouteDuration)).Err()
	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	var response schemas.GoongDirectionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	currentLocation := schemas.Point{
		Lat: optimizedPoints[0].Lat,
		Lng: optimizedPoints[0].Lng,
	}

	// Store the create ride request in the database
	rideRequestID, err := s.repo.CreateHitchRide(response, userID, currentLocation)

	if err != nil {
		return schemas.GoongDirectionsResponse{}, uuid.Nil, err
	}

	return response, rideRequestID, nil
}

func (s *MapService) GetGeoCode(ctx context.Context, point schemas.Point) (schemas.GoongReverseGeocodeResponse, error) {
	// Build the request URL
	baseURL, err := url.Parse(fmt.Sprintf("%s/geocode", s.cfg.GoongApiURL))
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("api_key", s.cfg.GoongAPIKey)

	location := fmt.Sprintf("%f,%f", point.Lat, point.Lng)
	params.Set("latlng", location)

	baseURL.RawQuery = params.Encode()
	url := baseURL.String()

	// Check the cache
	cacheKey := fmt.Sprintf("maps:geocode:%s", url)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var response schemas.GoongReverseGeocodeResponse
		if err := json.Unmarshal(cachedData, &response); err != nil {
			return schemas.GoongReverseGeocodeResponse{}, err
		}
		return response, nil
	}

	// If cache miss, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return schemas.GoongReverseGeocodeResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	// Cache the response
	err = s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCachePlaceDetailDuration)).Err()
	if err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	var response schemas.GoongReverseGeocodeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongReverseGeocodeResponse{}, err
	}

	return response, nil
}

// Make sure MapsService implements IMapsService
var _ IMapService = (*MapService)(nil)
