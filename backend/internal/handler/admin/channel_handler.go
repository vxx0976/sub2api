package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ChannelHandler handles admin channel management
type ChannelHandler struct {
	channelService *service.ChannelService
}

// NewChannelHandler creates a new admin channel handler
func NewChannelHandler(channelService *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// CreateChannelRequest represents create channel request
type CreateChannelRequest struct {
	Name           string            `json:"name" binding:"required,max=100"`
	Description    *string           `json:"description"`
	Platform       *string           `json:"platform"`
	Status         string            `json:"status"`
	IconURL        *string           `json:"icon_url"`
	WebsiteURL     *string           `json:"website_url"`
	BalanceURL     *string           `json:"balance_url"`
	BalanceMethod  string            `json:"balance_method"`
	BalanceHeaders map[string]string `json:"balance_headers"`
	BalanceBody    *string           `json:"balance_body"`
	BalancePath    *string           `json:"balance_path"`
	BalanceUnit    string            `json:"balance_unit"`
}

// UpdateChannelRequest represents update channel request
type UpdateChannelRequest struct {
	Name           *string           `json:"name" binding:"omitempty,max=100"`
	Description    *string           `json:"description"`
	Platform       *string           `json:"platform"`
	Status         *string           `json:"status"`
	IconURL        *string           `json:"icon_url"`
	WebsiteURL     *string           `json:"website_url"`
	BalanceURL     *string           `json:"balance_url"`
	BalanceMethod  *string           `json:"balance_method"`
	BalanceHeaders map[string]string `json:"balance_headers"`
	BalanceBody    *string           `json:"balance_body"`
	BalancePath    *string           `json:"balance_path"`
	BalanceUnit    *string           `json:"balance_unit"`
}

// List handles listing all channels with pagination and filters
// GET /api/v1/admin/channels
func (h *ChannelHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	platform := c.Query("platform")
	status := c.Query("status")
	search := c.Query("search")

	channels, pagination, err := h.channelService.List(c.Request.Context(), page, pageSize, platform, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.Channel, 0, len(channels))
	for i := range channels {
		out = append(out, *dto.ChannelFromService(&channels[i]))
	}
	response.PaginatedWithResult(c, out, toResponsePagination(pagination))
}

// GetByID handles getting a channel by ID
// GET /api/v1/admin/channels/:id
func (h *ChannelHandler) GetByID(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), channelID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ChannelFromService(channel))
}

// Create handles creating a new channel
// POST /api/v1/admin/channels
func (h *ChannelHandler) Create(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	channel, err := h.channelService.Create(c.Request.Context(), &service.CreateChannelInput{
		Name:           req.Name,
		Description:    req.Description,
		Platform:       req.Platform,
		Status:         req.Status,
		IconURL:        req.IconURL,
		WebsiteURL:     req.WebsiteURL,
		BalanceURL:     req.BalanceURL,
		BalanceMethod:  req.BalanceMethod,
		BalanceHeaders: req.BalanceHeaders,
		BalanceBody:    req.BalanceBody,
		BalancePath:    req.BalancePath,
		BalanceUnit:    req.BalanceUnit,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ChannelFromService(channel))
}

// Update handles updating a channel
// PUT /api/v1/admin/channels/:id
func (h *ChannelHandler) Update(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	channel, err := h.channelService.Update(c.Request.Context(), channelID, &service.UpdateChannelInput{
		Name:           req.Name,
		Description:    req.Description,
		Platform:       req.Platform,
		Status:         req.Status,
		IconURL:        req.IconURL,
		WebsiteURL:     req.WebsiteURL,
		BalanceURL:     req.BalanceURL,
		BalanceMethod:  req.BalanceMethod,
		BalanceHeaders: req.BalanceHeaders,
		BalanceBody:    req.BalanceBody,
		BalancePath:    req.BalancePath,
		BalanceUnit:    req.BalanceUnit,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ChannelFromService(channel))
}

// Delete handles deleting a channel
// DELETE /api/v1/admin/channels/:id
func (h *ChannelHandler) Delete(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	err = h.channelService.Delete(c.Request.Context(), channelID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Channel deleted successfully"})
}

// CheckBalance handles checking the balance of a channel
// POST /api/v1/admin/channels/:id/check-balance
func (h *ChannelHandler) CheckBalance(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	channel, err := h.channelService.CheckBalance(c.Request.Context(), channelID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ChannelFromService(channel))
}
