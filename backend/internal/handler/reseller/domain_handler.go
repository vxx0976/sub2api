package reseller

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// DomainHandler handles reseller domain management
type DomainHandler struct {
	resellerService      *service.ResellerService
	settingService       *service.SettingService
	invalidateDomainFunc func(domain string)
}

// NewDomainHandler creates a new DomainHandler
func NewDomainHandler(resellerService *service.ResellerService, settingService *service.SettingService) *DomainHandler {
	return &DomainHandler{resellerService: resellerService, settingService: settingService}
}

// SetCacheInvalidator sets the function to call when a domain's settings change
func (h *DomainHandler) SetCacheInvalidator(fn func(domain string)) {
	h.invalidateDomainFunc = fn
}

func (h *DomainHandler) invalidateDomainCache(domain string) {
	if h.invalidateDomainFunc != nil {
		h.invalidateDomainFunc(domain)
	}
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

	// Invalidate HTML cache for this domain so new settings take effect
	h.invalidateDomainCache(domain.Domain)

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

	domainName, err := h.resellerService.DeleteDomain(c.Request.Context(), resellerID, domainID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.invalidateDomainCache(domainName)

	response.Success(c, gin.H{"message": "Domain deleted successfully"})
}

// ServerInfo returns the main site's domain and resolved IP for DNS setup guidance
// GET /api/v1/reseller/domains/server-info
func (h *DomainHandler) ServerInfo(c *gin.Context) {
	// Prefer api_base_url from system settings (resolves to the public IP)
	host := ""
	if h.settingService != nil {
		settings, err := h.settingService.GetPublicSettings(c.Request.Context())
		if err == nil && settings.APIBaseURL != "" {
			if u, err := url.Parse(settings.APIBaseURL); err == nil && u.Hostname() != "" {
				host = u.Hostname()
			}
		}
	}
	// Fallback to request host
	if host == "" {
		host = c.Request.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
	}

	var serverIP string
	ips, err := net.LookupHost(host)
	if err == nil && len(ips) > 0 {
		// Prefer IPv4
		for _, ip := range ips {
			if !strings.Contains(ip, ":") {
				serverIP = ip
				break
			}
		}
		if serverIP == "" {
			serverIP = ips[0]
		}
	}

	response.Success(c, gin.H{
		"domain": host,
		"ip":     serverIP,
	})
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
