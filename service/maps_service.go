package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"time"

	"github.com/redis/go-redis/v9"
)

type IMapService interface {
	GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool) (schemas.GoongAutoCompleteResponse, error)
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
	err = s.redisClient.Set(ctx, cacheKey, body, time.Second*time.Duration(s.cfg.GoongCacheExpiredDuration)).Err()

	if err != nil {
		return schemas.GoongAutoCompleteResponse{}, err
	}

	var response schemas.GoongAutoCompleteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return schemas.GoongAutoCompleteResponse{}, err
	}

	return response, nil
}

// Make sure MapsService implements IMapsService
var _ IMapService = (*MapService)(nil)
