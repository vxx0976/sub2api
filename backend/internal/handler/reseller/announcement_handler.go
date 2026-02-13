package reseller

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AnnouncementHandler handles reseller announcement endpoints
type AnnouncementHandler struct {
	resellerService *service.ResellerService
}

// NewAnnouncementHandler creates a new AnnouncementHandler
func NewAnnouncementHandler(resellerService *service.ResellerService) *AnnouncementHandler {
	return &AnnouncementHandler{resellerService: resellerService}
}

// CreateAnnouncementRequest represents the request to create an announcement
type CreateAnnouncementRequest struct {
	Title    string     `json:"title" binding:"required"`
	Content  string     `json:"content" binding:"required"`
	Status   string     `json:"status"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

// UpdateAnnouncementRequest represents the request to update an announcement
type UpdateAnnouncementRequest struct {
	Title    *string     `json:"title"`
	Content  *string     `json:"content"`
	Status   *string     `json:"status"`
	StartsAt **time.Time `json:"starts_at"`
	EndsAt   **time.Time `json:"ends_at"`
}

// List returns announcements owned by the reseller
// GET /api/v1/reseller/announcements
func (h *AnnouncementHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	announcements, pag, err := h.resellerService.ListAnnouncements(c.Request.Context(), resellerID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.Announcement, 0, len(announcements))
	for i := range announcements {
		out = append(out, *dto.AnnouncementFromService(&announcements[i]))
	}

	response.PaginatedWithResult(c, out, &response.PaginationResult{
		Total:    pag.Total,
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Pages:    pag.Pages,
	})
}

// Create creates an announcement owned by the reseller
// POST /api/v1/reseller/announcements
func (h *AnnouncementHandler) Create(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := &service.CreateAnnouncementInput{
		Title:    req.Title,
		Content:  req.Content,
		Status:   req.Status,
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
	}

	a, err := h.resellerService.CreateAnnouncement(c.Request.Context(), resellerID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.AnnouncementFromService(a))
}

// Update updates an announcement owned by the reseller
// PUT /api/v1/reseller/announcements/:id
func (h *AnnouncementHandler) Update(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	var req UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := &service.UpdateAnnouncementInput{
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
	}

	a, err := h.resellerService.UpdateAnnouncement(c.Request.Context(), resellerID, announcementID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.AnnouncementFromService(a))
}

// Delete deletes an announcement owned by the reseller
// DELETE /api/v1/reseller/announcements/:id
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	if err := h.resellerService.DeleteAnnouncement(c.Request.Context(), resellerID, announcementID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, nil)
}
