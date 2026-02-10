package reseller

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// DomainHandler handles reseller domain management
type DomainHandler struct {
	resellerService *service.ResellerService
}

// NewDomainHandler creates a new DomainHandler
func NewDomainHandler(resellerService *service.ResellerService) *DomainHandler {
	return &DomainHandler{resellerService: resellerService}
}

// List lists the reseller's domains
// GET /api/v1/reseller/domains
func (h *DomainHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	page, pageSize := response.ParsePagination(c)

	domains, pagination, err := h.resellerService.ListDomains(c.Request.Context(), resellerID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.PaginatedWithResult(c, domains, &response.PaginationResult{
		Total:    pagination.Total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Pages:    pagination.Pages,
	})
}

// Create creates a new domain
// POST /api/v1/reseller/domains
func (h *DomainHandler) Create(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	var req service.CreateDomainInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	domain, err := h.resellerService.CreateDomain(c.Request.Context(), resellerID, &req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, domain)
}

// Update updates a domain
// PUT /api/v1/reseller/domains/:id
func (h *DomainHandler) Update(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid domain ID")
		return
	}

	var req service.UpdateDomainInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	domain, err := h.resellerService.UpdateDomain(c.Request.Context(), resellerID, domainID, &req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, domain)
}

// Delete deletes a domain
// DELETE /api/v1/reseller/domains/:id
func (h *DomainHandler) Delete(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid domain ID")
		return
	}

	if err := h.resellerService.DeleteDomain(c.Request.Context(), resellerID, domainID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Domain deleted successfully"})
}

// Verify verifies a domain via DNS TXT record
// POST /api/v1/reseller/domains/:id/verify
func (h *DomainHandler) Verify(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid domain ID")
		return
	}

	domain, err := h.resellerService.VerifyDomain(c.Request.Context(), resellerID, domainID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, domain)
}
