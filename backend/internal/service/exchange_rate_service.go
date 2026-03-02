package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// ExchangeRateService fetches and caches the USD/CNY exchange rate.
// Rate is fetched on demand and cached for 10 minutes.
type ExchangeRateService struct {
	httpClient *http.Client
	mu         sync.RWMutex
	cachedRate float64
	cachedAt   time.Time
	cacheTTL   time.Duration
}

func NewExchangeRateService() *ExchangeRateService {
	return &ExchangeRateService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: 10 * time.Minute,
	}
}

// GetUsdCnyRate returns the current USD/CNY rate.
// Uses a cached value if fresh enough, otherwise fetches from external API.
// Returns 0 if fetch fails and no cached value exists.
func (s *ExchangeRateService) GetUsdCnyRate() float64 {
	s.mu.RLock()
	if s.cachedRate > 0 && time.Since(s.cachedAt) < s.cacheTTL {
		rate := s.cachedRate
		s.mu.RUnlock()
		return rate
	}
	s.mu.RUnlock()

	rate, err := s.fetchUsdCnyRate()
	if err != nil {
		logger.LegacyPrintf("service.exchange_rate", "Failed to fetch USD/CNY rate: %v", err)
		// Return stale cache if available
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.cachedRate
	}

	// Sanity check
	if rate < 5.0 || rate > 10.0 {
		logger.LegacyPrintf("service.exchange_rate", "Fetched rate %.4f out of range [5, 10], using cached", rate)
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.cachedRate
	}

	s.mu.Lock()
	s.cachedRate = rate
	s.cachedAt = time.Now()
	s.mu.Unlock()

	logger.LegacyPrintf("service.exchange_rate", "Fetched USD/CNY rate: %.4f", rate)
	return rate
}

// exchangeRateResponse is the response from the exchange rate API.
type exchangeRateResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

func (s *ExchangeRateService) fetchUsdCnyRate() (float64, error) {
	// Primary: open.er-api.com (free, no key required)
	rate, err := s.fetchFromErApi()
	if err == nil {
		return rate, nil
	}
	logger.LegacyPrintf("service.exchange_rate", "Primary API failed: %v, trying fallback", err)

	// Fallback: exchangerate-api.com
	rate, err = s.fetchFromExchangeRateApi()
	if err == nil {
		return rate, nil
	}

	return 0, fmt.Errorf("all exchange rate APIs failed: %w", err)
}

func (s *ExchangeRateService) fetchFromErApi() (float64, error) {
	resp, err := s.httpClient.Get("https://open.er-api.com/v6/latest/USD")
	if err != nil {
		return 0, fmt.Errorf("er-api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return 0, fmt.Errorf("er-api read body: %w", err)
	}

	var result exchangeRateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("er-api parse: %w", err)
	}

	rate, ok := result.Rates["CNY"]
	if !ok || rate <= 0 {
		return 0, fmt.Errorf("er-api: CNY rate not found")
	}
	return rate, nil
}

func (s *ExchangeRateService) fetchFromExchangeRateApi() (float64, error) {
	resp, err := s.httpClient.Get("https://api.exchangerate-api.com/v4/latest/USD")
	if err != nil {
		return 0, fmt.Errorf("exchangerate-api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return 0, fmt.Errorf("exchangerate-api read body: %w", err)
	}

	var result exchangeRateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("exchangerate-api parse: %w", err)
	}

	rate, ok := result.Rates["CNY"]
	if !ok || rate <= 0 {
		return 0, fmt.Errorf("exchangerate-api: CNY rate not found")
	}
	return rate, nil
}
