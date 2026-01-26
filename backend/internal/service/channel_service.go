package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// ChannelService handles channel business logic
type ChannelService struct {
	channelRepo ChannelRepository
	httpClient  *http.Client
}

// NewChannelService creates a new channel service
func NewChannelService(channelRepo ChannelRepository) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateChannelInput represents input for creating a channel
type CreateChannelInput struct {
	Name           string
	Description    *string
	Platform       *string
	Status         string
	IconURL        *string
	WebsiteURL     *string
	BalanceURL     *string
	BalanceMethod  string
	BalanceHeaders map[string]string
	BalanceBody    *string
	BalancePath    *string
	BalanceUnit    string
}

// UpdateChannelInput represents input for updating a channel
type UpdateChannelInput struct {
	Name           *string
	Description    *string
	Platform       *string
	Status         *string
	IconURL        *string
	WebsiteURL     *string
	BalanceURL     *string
	BalanceMethod  *string
	BalanceHeaders map[string]string
	BalanceBody    *string
	BalancePath    *string
	BalanceUnit    *string
}

// List returns a paginated list of channels
func (s *ChannelService) List(ctx context.Context, page, pageSize int, platform, status, search string) ([]Channel, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.channelRepo.List(ctx, params, platform, status, search)
}

// GetByID returns a channel by ID
func (s *ChannelService) GetByID(ctx context.Context, id int64) (*Channel, error) {
	return s.channelRepo.GetByID(ctx, id)
}

// Create creates a new channel
func (s *ChannelService) Create(ctx context.Context, input *CreateChannelInput) (*Channel, error) {
	// Set defaults
	status := input.Status
	if status == "" {
		status = ChannelStatusActive
	}
	balanceMethod := input.BalanceMethod
	if balanceMethod == "" {
		balanceMethod = "GET"
	}
	balanceUnit := input.BalanceUnit
	if balanceUnit == "" {
		balanceUnit = "$"
	}

	channel := &Channel{
		Name:           input.Name,
		Description:    input.Description,
		Platform:       input.Platform,
		Status:         status,
		IconURL:        input.IconURL,
		WebsiteURL:     input.WebsiteURL,
		BalanceURL:     input.BalanceURL,
		BalanceMethod:  balanceMethod,
		BalanceHeaders: input.BalanceHeaders,
		BalanceBody:    input.BalanceBody,
		BalancePath:    input.BalancePath,
		BalanceUnit:    balanceUnit,
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// Update updates an existing channel
func (s *ChannelService) Update(ctx context.Context, id int64, input *UpdateChannelInput) (*Channel, error) {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Name != nil {
		channel.Name = *input.Name
	}
	if input.Description != nil {
		channel.Description = input.Description
	}
	if input.Platform != nil {
		channel.Platform = input.Platform
	}
	if input.Status != nil {
		channel.Status = *input.Status
	}
	if input.IconURL != nil {
		channel.IconURL = input.IconURL
	}
	if input.WebsiteURL != nil {
		channel.WebsiteURL = input.WebsiteURL
	}
	if input.BalanceURL != nil {
		channel.BalanceURL = input.BalanceURL
	}
	if input.BalanceMethod != nil {
		channel.BalanceMethod = *input.BalanceMethod
	}
	if input.BalanceHeaders != nil {
		channel.BalanceHeaders = input.BalanceHeaders
	}
	if input.BalanceBody != nil {
		channel.BalanceBody = input.BalanceBody
	}
	if input.BalancePath != nil {
		channel.BalancePath = input.BalancePath
	}
	if input.BalanceUnit != nil {
		channel.BalanceUnit = *input.BalanceUnit
	}

	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// Delete deletes a channel by ID
func (s *ChannelService) Delete(ctx context.Context, id int64) error {
	return s.channelRepo.Delete(ctx, id)
}

// CheckBalance queries the balance API and updates the cached balance
func (s *ChannelService) CheckBalance(ctx context.Context, id int64) (*Channel, error) {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// If no balance URL configured, just return the channel
	if channel.BalanceURL == nil || *channel.BalanceURL == "" {
		return channel, nil
	}

	// Build request
	var req *http.Request
	if channel.BalanceMethod == "POST" && channel.BalanceBody != nil {
		req, err = http.NewRequestWithContext(ctx, "POST", *channel.BalanceURL, strings.NewReader(*channel.BalanceBody))
	} else {
		req, err = http.NewRequestWithContext(ctx, "GET", *channel.BalanceURL, nil)
	}
	if err != nil {
		lastError := fmt.Sprintf("failed to create request: %v", err)
		_ = s.channelRepo.UpdateBalance(ctx, id, nil, time.Now(), &lastError)
		channel.LastError = &lastError
		channel.LastCheckAt = ptrTime(time.Now())
		channel.Status = ChannelStatusError
		return channel, nil
	}

	// Set headers
	if channel.BalanceHeaders != nil {
		for k, v := range channel.BalanceHeaders {
			req.Header.Set(k, v)
		}
	}
	if req.Header.Get("Content-Type") == "" && channel.BalanceMethod == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		lastError := fmt.Sprintf("request failed: %v", err)
		_ = s.channelRepo.UpdateBalance(ctx, id, nil, time.Now(), &lastError)
		channel.LastError = &lastError
		channel.LastCheckAt = ptrTime(time.Now())
		channel.Status = ChannelStatusError
		return channel, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		lastError := fmt.Sprintf("failed to read response: %v", err)
		_ = s.channelRepo.UpdateBalance(ctx, id, nil, time.Now(), &lastError)
		channel.LastError = &lastError
		channel.LastCheckAt = ptrTime(time.Now())
		channel.Status = ChannelStatusError
		return channel, nil
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		lastError := fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
		_ = s.channelRepo.UpdateBalance(ctx, id, nil, time.Now(), &lastError)
		channel.LastError = &lastError
		channel.LastCheckAt = ptrTime(time.Now())
		channel.Status = ChannelStatusError
		return channel, nil
	}

	// Parse response and extract balance
	balance, err := s.extractBalance(body, channel.BalancePath)
	if err != nil {
		lastError := fmt.Sprintf("failed to extract balance: %v", err)
		_ = s.channelRepo.UpdateBalance(ctx, id, nil, time.Now(), &lastError)
		channel.LastError = &lastError
		channel.LastCheckAt = ptrTime(time.Now())
		channel.Status = ChannelStatusError
		return channel, nil
	}

	// Update cached balance
	now := time.Now()
	if err := s.channelRepo.UpdateBalance(ctx, id, &balance, now, nil); err != nil {
		return nil, err
	}

	channel.CachedBalance = &balance
	channel.LastCheckAt = &now
	channel.LastError = nil
	channel.Status = ChannelStatusActive

	return channel, nil
}

// extractBalance extracts the balance value from JSON response using the given path
func (s *ChannelService) extractBalance(body []byte, path *string) (float64, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("invalid JSON response: %v", err)
	}

	// If no path specified, try to parse the whole response as a number
	if path == nil || *path == "" {
		return parseFloat(data)
	}

	// Navigate the path
	parts := strings.Split(*path, ".")
	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return 0, fmt.Errorf("path '%s' not found in response", *path)
			}
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return 0, fmt.Errorf("invalid array index '%s' in path", part)
			}
			current = v[idx]
		default:
			return 0, fmt.Errorf("cannot navigate path '%s': unexpected type at '%s'", *path, part)
		}
	}

	return parseFloat(current)
}

// parseFloat converts various types to float64
func parseFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse '%s' as number", val)
		}
		return f, nil
	case json.Number:
		return val.Float64()
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
