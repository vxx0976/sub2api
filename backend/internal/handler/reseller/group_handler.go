package reseller

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// GroupHandler handles reseller group (package) management
type GroupHandler struct {
	resellerService *service.ResellerService
}

// NewGroupHandler creates a new GroupHandler
func NewGroupHandler(resellerService *service.ResellerService) *GroupHandler {
	return &GroupHandler{resellerService: resellerService}
}

// ListTemplates returns admin groups marked as reseller templates
// GET /api/v1/reseller/groups/templates
func (h *GroupHandler) ListTemplates(c *gin.Context) {
	templates, err := h.resellerService.ListTemplates(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, templates)
}

// List lists groups owned by the reseller
// GET /api/v1/reseller/groups
func (h *GroupHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	page, pageSize := response.ParsePagination(c)

	groups, pagination, err := h.resellerService.ListGroups(c.Request.Context(), resellerID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.PaginatedWithResult(c, groups, &response.PaginationResult{
		Total:    pagination.Total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Pages:    pagination.Pages,
	})
}

// GetByID returns a group owned by the reseller
// GET /api/v1/reseller/groups/:id
func (h *GroupHandler) GetByID(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	group, err := h.resellerService.GetGroup(c.Request.Context(), resellerID, groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, group)
}

// Create creates a new group for the reseller
// POST /api/v1/reseller/groups
func (h *GroupHandler) Create(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	var req service.CreateResellerGroupInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	group, err := h.resellerService.CreateGroup(c.Request.Context(), resellerID, &req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, group)
}

// Update updates a group owned by the reseller
// PUT /api/v1/reseller/groups/:id
func (h *GroupHandler) Update(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	var req service.UpdateResellerGroupInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	group, err := h.resellerService.UpdateGroup(c.Request.Context(), resellerID, groupID, &req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, group)
}

// Delete deletes a group owned by the reseller
// DELETE /api/v1/reseller/groups/:id
func (h *GroupHandler) Delete(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	if err := h.resellerService.DeleteGroup(c.Request.Context(), resellerID, groupID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Group deleted successfully"})
}
