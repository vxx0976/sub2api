package reseller

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler handles reseller user list endpoints
type UserHandler struct {
	resellerService *service.ResellerService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(resellerService *service.ResellerService) *UserHandler {
	return &UserHandler{resellerService: resellerService}
}

// List returns users belonging to the reseller (parent_id = resellerID)
// GET /api/v1/reseller/users
func (h *UserHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}

	filters := service.UserListFilters{
		ParentID: &resellerID,
		Search:   search,
	}

	users, pag, err := h.resellerService.ListUsers(c.Request.Context(), resellerID, page, pageSize, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.User, 0, len(users))
	for i := range users {
		out = append(out, dto.UserFromServiceShallow(&users[i]))
	}

	response.PaginatedWithResult(c, out, &response.PaginationResult{
		Total:    pag.Total,
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Pages:    pag.Pages,
	})
}
