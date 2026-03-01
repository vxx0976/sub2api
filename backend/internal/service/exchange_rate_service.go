package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// ExchangeRateService periodically fetches the USD/CNY exchange rate
// and updates the recharge config.
type ExchangeRateService struct {
	settingService *SettingService
	httpClient     *http.Client
	stopCh         chan struct{}
	mu             sync.Mutex
	lastRate       float64
}

func NewExchangeRateService(settingService *SettingService) *ExchangeRateService {
	return &ExchangeRateService{
		settingService: settingService,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

// Start begins the periodic exchange rate update (every 6 hours).
func (s *ExchangeRateService) Start() {
	go s.run()
}

func (s *ExchangeRateService) Stop() {
	close(s.stopCh)
}

func (s *ExchangeRateService) run() {
	// Fetch immediately on startup.
	s.updateRate()

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateRate()
		case <-s.stopCh:
			return
		}
	}
}

func (s *ExchangeRateService) updateRate() {
	rate, err := s.fetchUsdCnyRate()
	if err != nil {
		logger.LegacyPrintf("service.exchange_rate", "Failed to fetch USD/CNY rate: %v", err)
		return
	}

	// Sanity check: rate should be in a reasonable range.
	if rate < 5.0 || rate > 10.0 {
		logger.LegacyPrintf("service.exchange_rate", "Fetched rate %.4f is out of reasonable range [5, 10], skipping", rate)
		return
	}

	s.mu.Lock()
	s.lastRate = rate
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := s.settingService.GetRechargeConfig(ctx)
	if err != nil {
		logger.LegacyPrintf("service.exchange_rate", "Failed to get recharge config: %v", err)
		return
	}

	config.UsdCnyRate = rate
	if err := s.settingService.UpdateRechargeConfig(ctx, config); err != nil {
		logger.LegacyPrintf("service.exchange_rate", "Failed to update recharge config with rate %.4f: %v", rate, err)
		return
	}

	logger.LegacyPrintf("service.exchange_rate", "Updated USD/CNY rate to %.4f", rate)
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
