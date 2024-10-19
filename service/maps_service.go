package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"shareway/repository"
	"shareway/util"
	"time"

	"github.com/redis/go-redis/v9"
)

type IMapsService interface {
	GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool) ([]string, error)
}

type MapsService struct {
	repo        repository.IMapsRepository
	cfg         util.Config
	redisClient *redis.Client
}

func NewMapsService(repo repository.IMapsRepository, cfg util.Config, redisClient *redis.Client) IMapsService {
	return &MapsService{
		repo:        repo,
		cfg:         cfg,
		redisClient: redisClient,
	}
}

func (s *MapsService) GetAutoComplete(ctx context.Context, input string, limit int, location string, radius int, moreCompound bool) ([]string, error) {
	// Build the request URL
	url := fmt.Sprintf("%s/place/autocomplete?apikey=%s&input=%s", s.cfg.GoongApiURL, s.cfg.GoongAPIKey, input)

	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	if location != "" {
		url += fmt.Sprintf("&location=%s", location)
	}

	if radius > 0 {
		url += fmt.Sprintf("&radius=%d", radius)
	}

	if moreCompound {
		url += "&more_compound=true"
	}

	// Check cache for partial matches

	// Check the cache
	cacheKey := fmt.Sprintf("maps:autocomplete:%s", url)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		// If cache hit, return the cached data
		return []string{string(cachedData)}, nil
	}

	// If cache miss, fetch data from Goong API
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache the response
	err = s.redisClient.Set(ctx, cacheKey, body, time.Duration(s.cfg.GoongCacheExpiredDuration)).Err()
	if err != nil {
		return nil, err
	}

	return []string{string(body)}, nil
}

// Make sure MapsService implements IMapsService
var _ IMapsService = (*MapsService)(nil)
