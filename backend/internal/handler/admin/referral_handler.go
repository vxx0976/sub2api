package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// ReferralHandler handles admin referral management requests
type ReferralHandler struct {
	referralService *service.ReferralService
}

// NewReferralHandler creates a new admin ReferralHandler
func NewReferralHandler(referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{
		referralService: referralService,
	}
}

// AdminReferralRecord represents a referral record for admin view
type AdminReferralRecord struct {
	ID                 int64    `json:"id"`
	ReferrerID         int64    `json:"referrer_id"`
	ReferrerEmail      string   `json:"referrer_email"`
	InviteeID          int64    `json:"invitee_id"`
	InviteeEmail       string   `json:"invitee_email"`
	TriggerOrderID     int64    `json:"trigger_order_id"`
	ReferrerReward     float64  `json:"referrer_reward"`
	InviteeReward      float64  `json:"invitee_reward"`
	SkipReferrerReason *string  `json:"skip_referrer_reason,omitempty"`
	Status             string   `json:"status"`
	CreatedAt          string   `json:"created_at"`
}

// AdminReferralListResponse response for admin referral list
type AdminReferralListResponse struct {
	Items      []AdminReferralRecord `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// List gets all referral records for admin
// GET /api/v1/admin/referrals
func (h *ReferralHandler) List(c *gin.Context) {
	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse search param
	search := c.Query("search")

	records, paginationResult, err := h.referralService.GetAllReferralRecords(c.Request.Context(), page, pageSize, search)
	if err != nil {
		if !response.ErrorFrom(c, err) {
			response.InternalError(c, err.Error())
		}
		return
	}

	items := make([]AdminReferralRecord, len(records))
	for i, rec := range records {
		items[i] = AdminReferralRecord{
			ID:                 rec.ID,
			ReferrerID:         rec.ReferrerID,
			ReferrerEmail:      rec.ReferrerEmail,
			InviteeID:          rec.InviteeID,
			InviteeEmail:       rec.InviteeEmail,
			TriggerOrderID:     rec.TriggerOrderID,
			ReferrerReward:     rec.ReferrerReward,
			InviteeReward:      rec.InviteeReward,
			SkipReferrerReason: rec.SkipReferrerReason,
			Status:             rec.Status,
			CreatedAt:          rec.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Success(c, AdminReferralListResponse{
		Items:      items,
		Total:      int(paginationResult.Total),
		Page:       paginationResult.Page,
		PageSize:   paginationResult.PageSize,
		TotalPages: paginationResult.Pages,
	})
}

// GetStats gets referral statistics for admin
// GET /api/v1/admin/referrals/stats
func (h *ReferralHandler) GetStats(c *gin.Context) {
	stats, err := h.referralService.GetAdminReferralStats(c.Request.Context())
	if err != nil {
		if !response.ErrorFrom(c, err) {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, stats)
}
