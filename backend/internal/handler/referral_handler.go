package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// ReferralHandler handles referral-related requests
type ReferralHandler struct {
	referralService *service.ReferralService
}

// NewReferralHandler creates a new ReferralHandler
func NewReferralHandler(referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{
		referralService: referralService,
	}
}

// GetCodeResponse response for getting referral code
type GetCodeResponse struct {
	Code string `json:"code"`
}

// GetCode gets or generates the user's referral code
// GET /api/v1/referral/code
func (h *ReferralHandler) GetCode(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := subject.UserID

	code, err := h.referralService.GetUserReferralCode(c.Request.Context(), userID)
	if err != nil {
		if !response.ErrorFrom(c, err) {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, GetCodeResponse{Code: code})
}

// StatsResponse response for getting referral stats
type StatsResponse struct {
	TotalInvited   int     `json:"total_invited"`
	TotalRewarded  int     `json:"total_rewarded"`
	PendingPayment int     `json:"pending_payment"`
	TotalEarnings  float64 `json:"total_earnings"`
}

// GetStats gets referral statistics for the current user
// GET /api/v1/referral/stats
func (h *ReferralHandler) GetStats(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := subject.UserID

	stats, err := h.referralService.GetReferralStats(c.Request.Context(), userID)
	if err != nil {
		if !response.ErrorFrom(c, err) {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, StatsResponse{
		TotalInvited:   stats.TotalInvited,
		TotalRewarded:  stats.TotalRewarded,
		PendingPayment: stats.PendingPayment,
		TotalEarnings:  stats.TotalEarnings,
	})
}

// InviteeResponse response item for invitee list
type InviteeResponse struct {
	ID              int64    `json:"id"`
	Email           string   `json:"email"`
	Status          string   `json:"status"`
	JoinedAt        string   `json:"joined_at"`
	RewardedAt      *string  `json:"rewarded_at,omitempty"`
	RewardAmount    *float64 `json:"reward_amount,omitempty"`
	ReferrerEarning *float64 `json:"referrer_earning,omitempty"`
}

// InviteesListResponse response for getting invitees list
type InviteesListResponse struct {
	Items      []InviteeResponse `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// GetInvitees gets the list of invitees for the current user
// GET /api/v1/referral/invitees
func (h *ReferralHandler) GetInvitees(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	userID := subject.UserID

	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	invitees, paginationResult, err := h.referralService.GetInvitees(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		if !response.ErrorFrom(c, err) {
			response.InternalError(c, err.Error())
		}
		return
	}

	items := make([]InviteeResponse, len(invitees))
	for i, inv := range invitees {
		item := InviteeResponse{
			ID:              inv.ID,
			Email:           inv.Email,
			Status:          inv.Status,
			JoinedAt:        inv.JoinedAt.Format("2006-01-02 15:04:05"),
			RewardAmount:    inv.RewardAmount,
			ReferrerEarning: inv.ReferrerEarning,
		}
		if inv.RewardedAt != nil {
			formatted := inv.RewardedAt.Format("2006-01-02 15:04:05")
			item.RewardedAt = &formatted
		}
		items[i] = item
	}

	response.Success(c, InviteesListResponse{
		Items:      items,
		Total:      int(paginationResult.Total),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalPages: paginationResult.Pages,
	})
}
