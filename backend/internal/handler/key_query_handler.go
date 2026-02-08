package handler

import (
	"math"
	"sort"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// KeyQueryHandler handles public API key query requests (no JWT required)
type KeyQueryHandler struct {
	apiKeyService       *service.APIKeyService
	subscriptionService *service.SubscriptionService
	usageService        *service.UsageService
}

// NewKeyQueryHandler creates a new KeyQueryHandler
func NewKeyQueryHandler(
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	usageService *service.UsageService,
) *KeyQueryHandler {
	return &KeyQueryHandler{
		apiKeyService:       apiKeyService,
		subscriptionService: subscriptionService,
		usageService:        usageService,
	}
}

// keyQueryRequest is the request body for key query
type keyQueryRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// keyQueryKeyInfo represents API key information in the response
type keyQueryKeyInfo struct {
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	Quota          *float64   `json:"quota"`
	QuotaUsed      *float64   `json:"quota_used"`
	QuotaRemaining *float64   `json:"quota_remaining"`
	ExpiresAt      *time.Time `json:"expires_at"`
	DaysUntilExpiry *int      `json:"days_until_expiry"`
}

// keyQuerySubscriptionInfo represents subscription information in the response
type keyQuerySubscriptionInfo struct {
	GroupName       string     `json:"group_name"`
	Status          string     `json:"status"`
	ExpiresAt       *time.Time `json:"expires_at"`
	DaysRemaining   *int       `json:"days_remaining"`
	DailyLimitUSD   *float64   `json:"daily_limit_usd"`
	DailyUsageUSD   *float64   `json:"daily_usage_usd"`
	WeeklyLimitUSD  *float64   `json:"weekly_limit_usd"`
	WeeklyUsageUSD  *float64   `json:"weekly_usage_usd"`
	MonthlyLimitUSD *float64   `json:"monthly_limit_usd"`
	MonthlyUsageUSD *float64   `json:"monthly_usage_usd"`
}

// keyQueryTopModel represents a top model in usage summary
type keyQueryTopModel struct {
	Model string  `json:"model"`
	Count int64   `json:"count"`
	Cost  float64 `json:"cost"`
}

// keyQueryUsageSummary represents the 7-day usage summary
type keyQueryUsageSummary struct {
	TotalCost7d    float64            `json:"total_cost_7d"`
	RequestCount7d int64              `json:"request_count_7d"`
	TopModels      []keyQueryTopModel `json:"top_models"`
}

// keyQueryResponse is the full response for key query
type keyQueryResponse struct {
	Key          *keyQueryKeyInfo          `json:"key"`
	Subscription *keyQuerySubscriptionInfo `json:"subscription"`
	UsageSummary *keyQueryUsageSummary     `json:"usage_summary"`
}

// QueryKey handles POST /api/v1/public/key-query
func (h *KeyQueryHandler) QueryKey(c *gin.Context) {
	var req keyQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: api_key is required")
		return
	}

	ctx := c.Request.Context()

	// Look up the API key (uses repo.GetByKey which filters soft-deleted)
	apiKey, err := h.apiKeyService.GetByKey(ctx, req.APIKey)
	if err != nil {
		response.BadRequest(c, "Invalid API key")
		return
	}

	// Build key info
	keyInfo := &keyQueryKeyInfo{
		Name:   apiKey.Name,
		Status: apiKey.Status,
	}

	// Quota fields: nil if unlimited (Quota <= 0)
	if apiKey.Quota > 0 {
		quota := apiKey.Quota
		quotaUsed := apiKey.QuotaUsed
		remaining := apiKey.GetQuotaRemaining()
		keyInfo.Quota = &quota
		keyInfo.QuotaUsed = &quotaUsed
		keyInfo.QuotaRemaining = &remaining
	}

	// Expiry fields: nil if never expires
	if apiKey.ExpiresAt != nil {
		keyInfo.ExpiresAt = apiKey.ExpiresAt
		days := apiKey.GetDaysUntilExpiry()
		keyInfo.DaysUntilExpiry = &days
	}

	// Build subscription info (if key has a group and the user has an active subscription for it)
	var subInfo *keyQuerySubscriptionInfo
	if apiKey.GroupID != nil {
		sub, err := h.subscriptionService.GetActiveSubscription(ctx, apiKey.UserID, *apiKey.GroupID)
		if err == nil && sub != nil {
			subInfo = &keyQuerySubscriptionInfo{
				Status: sub.Status,
			}

			expiresAt := sub.ExpiresAt
			subInfo.ExpiresAt = &expiresAt
			daysRemaining := sub.DaysRemaining()
			subInfo.DaysRemaining = &daysRemaining

			// Get group info for limits
			if sub.Group != nil {
				subInfo.GroupName = sub.Group.Name
				if sub.Group.HasDailyLimit() {
					subInfo.DailyLimitUSD = sub.Group.DailyLimitUSD
					usage := sub.DailyUsageUSD
					subInfo.DailyUsageUSD = &usage
				}
				if sub.Group.HasWeeklyLimit() {
					subInfo.WeeklyLimitUSD = sub.Group.WeeklyLimitUSD
					usage := sub.WeeklyUsageUSD
					subInfo.WeeklyUsageUSD = &usage
				}
				if sub.Group.HasMonthlyLimit() {
					subInfo.MonthlyLimitUSD = sub.Group.MonthlyLimitUSD
					usage := sub.MonthlyUsageUSD
					subInfo.MonthlyUsageUSD = &usage
				}
			}
		}
	}

	// Build 7-day usage summary
	var usageSummary *keyQueryUsageSummary
	now := time.Now()
	startTime := now.AddDate(0, 0, -7)

	stats, err := h.usageService.GetStatsByUser(ctx, apiKey.UserID, startTime, now)
	if err == nil && stats != nil {
		usageSummary = &keyQueryUsageSummary{
			TotalCost7d:    math.Round(stats.TotalActualCost*100) / 100,
			RequestCount7d: stats.TotalRequests,
			TopModels:      make([]keyQueryTopModel, 0),
		}

		// Get top models
		modelStats, err := h.usageService.GetUserModelStats(ctx, apiKey.UserID, startTime, now)
		if err == nil && len(modelStats) > 0 {
			// Sort by request count descending
			sort.Slice(modelStats, func(i, j int) bool {
				return modelStats[i].Requests > modelStats[j].Requests
			})

			// Take top 5
			limit := 5
			if len(modelStats) < limit {
				limit = len(modelStats)
			}
			for _, ms := range modelStats[:limit] {
				usageSummary.TopModels = append(usageSummary.TopModels, keyQueryTopModel{
					Model: ms.Model,
					Count: ms.Requests,
					Cost:  math.Round(ms.ActualCost*100) / 100,
				})
			}
		}
	}

	resp := keyQueryResponse{
		Key:          keyInfo,
		Subscription: subInfo,
		UsageSummary: usageSummary,
	}

	response.Success(c, resp)
}

// validateAPIKey validates the API key from the request body and returns the key object.
// Returns nil and false if validation fails (response already written).
func (h *KeyQueryHandler) validateAPIKey(c *gin.Context) (*service.APIKey, bool) {
	var req keyQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: api_key is required")
		return nil, false
	}

	apiKey, err := h.apiKeyService.GetByKey(c.Request.Context(), req.APIKey)
	if err != nil {
		response.BadRequest(c, "Invalid API key")
		return nil, false
	}

	return apiKey, true
}

// keyUsageRequest is the request body for key usage endpoints
type keyUsageRequest struct {
	APIKey    string `json:"api_key" binding:"required"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Timezone  string `json:"timezone"`
	Model     string `json:"model"`
}

// parseKeyUsageDates parses start_date and end_date from the request body.
// Returns the time range. If dates are not provided, defaults to last 7 days.
func parseKeyUsageDates(req *keyUsageRequest) (startTime, endTime time.Time, errMsg string) {
	userTZ := req.Timezone

	if req.StartDate != "" && req.EndDate != "" {
		var err error
		startTime, err = timezone.ParseInUserLocation("2006-01-02", req.StartDate, userTZ)
		if err != nil {
			return time.Time{}, time.Time{}, "Invalid start_date format, use YYYY-MM-DD"
		}
		endTime, err = timezone.ParseInUserLocation("2006-01-02", req.EndDate, userTZ)
		if err != nil {
			return time.Time{}, time.Time{}, "Invalid end_date format, use YYYY-MM-DD"
		}
		endTime = endTime.Add(24*time.Hour - time.Nanosecond)
	} else {
		now := timezone.NowInUserLocation(userTZ)
		startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
		endTime = now
	}

	return startTime, endTime, ""
}

// ListUsage handles POST /api/v1/public/key-usage
func (h *KeyQueryHandler) ListUsage(c *gin.Context) {
	var req keyUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: api_key is required")
		return
	}

	ctx := c.Request.Context()

	apiKey, err := h.apiKeyService.GetByKey(ctx, req.APIKey)
	if err != nil {
		response.BadRequest(c, "Invalid API key")
		return
	}

	startTime, endTime, errMsg := parseKeyUsageDates(&req)
	if errMsg != "" {
		response.BadRequest(c, errMsg)
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	filters := usagestats.UsageLogFilters{
		APIKeyID:  apiKey.ID,
		Model:     req.Model,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	records, result, err := h.usageService.ListWithFilters(ctx, params, filters)
	if err != nil {
		response.BadRequest(c, "Failed to fetch usage logs")
		return
	}

	out := make([]dto.UsageLog, 0, len(records))
	for i := range records {
		out = append(out, *dto.UsageLogFromService(&records[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// UsageStats handles POST /api/v1/public/key-usage/stats
func (h *KeyQueryHandler) UsageStats(c *gin.Context) {
	var req keyUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: api_key is required")
		return
	}

	ctx := c.Request.Context()

	apiKey, err := h.apiKeyService.GetByKey(ctx, req.APIKey)
	if err != nil {
		response.BadRequest(c, "Invalid API key")
		return
	}

	startTime, endTime, errMsg := parseKeyUsageDates(&req)
	if errMsg != "" {
		response.BadRequest(c, errMsg)
		return
	}

	stats, err := h.usageService.GetStatsByAPIKey(ctx, apiKey.ID, startTime, endTime)
	if err != nil {
		response.BadRequest(c, "Failed to fetch usage stats")
		return
	}

	response.Success(c, stats)
}

// UsageModels handles POST /api/v1/public/key-usage/models
func (h *KeyQueryHandler) UsageModels(c *gin.Context) {
	var req keyUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: api_key is required")
		return
	}

	ctx := c.Request.Context()

	apiKey, err := h.apiKeyService.GetByKey(ctx, req.APIKey)
	if err != nil {
		response.BadRequest(c, "Invalid API key")
		return
	}

	startTime, endTime, errMsg := parseKeyUsageDates(&req)
	if errMsg != "" {
		response.BadRequest(c, errMsg)
		return
	}

	models, err := h.usageService.GetAPIKeyModelStats(ctx, apiKey.ID, startTime, endTime)
	if err != nil {
		response.BadRequest(c, "Failed to fetch model stats")
		return
	}

	response.Success(c, gin.H{
		"models":     models,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Format("2006-01-02"),
	})
}

// UsageTrend handles POST /api/v1/public/key-usage/trend
func (h *KeyQueryHandler) UsageTrend(c *gin.Context) {
	var req keyUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: api_key is required")
		return
	}

	ctx := c.Request.Context()

	apiKey, err := h.apiKeyService.GetByKey(ctx, req.APIKey)
	if err != nil {
		response.BadRequest(c, "Invalid API key")
		return
	}

	startTime, endTime, errMsg := parseKeyUsageDates(&req)
	if errMsg != "" {
		response.BadRequest(c, errMsg)
		return
	}

	trend, err := h.usageService.GetAPIKeyUsageTrend(ctx, apiKey.ID, startTime, endTime, "day")
	if err != nil {
		response.BadRequest(c, "Failed to fetch usage trend")
		return
	}

	response.Success(c, gin.H{
		"trend":      trend,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Format("2006-01-02"),
	})
}
