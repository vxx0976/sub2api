package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrChannelNotFound      = infraerrors.NotFound("CHANNEL_NOT_FOUND", "channel not found")
	ErrChannelAlreadyExists = infraerrors.Conflict("CHANNEL_ALREADY_EXISTS", "channel with this name already exists")
)

// Channel status constants
const (
	ChannelStatusActive   = "active"
	ChannelStatusInactive = "inactive"
	ChannelStatusError    = "error"
)

// Channel represents a channel configuration
type Channel struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Platform    *string `json:"platform,omitempty"`
	Status      string  `json:"status"`
	IconURL     *string `json:"icon_url,omitempty"`
	WebsiteURL  *string `json:"website_url,omitempty"`

	// 余额查询配置
	BalanceURL     *string           `json:"balance_url,omitempty"`
	BalanceMethod  string            `json:"balance_method"`
	BalanceHeaders map[string]string `json:"balance_headers,omitempty"`
	BalanceBody    *string           `json:"balance_body,omitempty"`
	BalancePath    *string           `json:"balance_path,omitempty"`
	BalanceUnit    string            `json:"balance_unit"`

	// 缓存的余额信息
	CachedBalance *float64   `json:"cached_balance,omitempty"`
	LastCheckAt   *time.Time `json:"last_check_at,omitempty"`
	LastError     *string    `json:"last_error,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsActive checks if the channel is active
func (c *Channel) IsActive() bool {
	return c.Status == ChannelStatusActive
}

// HasError checks if the channel has an error status
func (c *Channel) HasError() bool {
	return c.Status == ChannelStatusError
}

// ChannelRepository defines the interface for channel persistence
type ChannelRepository interface {
	Create(ctx context.Context, channel *Channel) error
	GetByID(ctx context.Context, id int64) (*Channel, error)
	GetByName(ctx context.Context, name string) (*Channel, error)
	Update(ctx context.Context, channel *Channel) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params pagination.PaginationParams, platform string, status string, search string) ([]Channel, *pagination.PaginationResult, error)
	UpdateBalance(ctx context.Context, id int64, balance *float64, lastCheckAt time.Time, lastError *string) error
}
