package reseller

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RedeemHandler handles reseller redeem code endpoints
type RedeemHandler struct {
	resellerService *service.ResellerService
}

// NewRedeemHandler creates a new RedeemHandler
func NewRedeemHandler(resellerService *service.ResellerService) *RedeemHandler {
	return &RedeemHandler{resellerService: resellerService}
}

// GenerateRequest represents the request to generate redeem codes
type GenerateRequest struct {
	Count int     `json:"count" binding:"required,min=1,max=100"`
	Value float64 `json:"value" binding:"required,gt=0"`
}

// Generate creates redeem codes funded from the reseller's balance
// POST /api/v1/reseller/redeem/generate
func (h *RedeemHandler) Generate(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	codes, err := h.resellerService.GenerateRedeemCodes(c.Request.Context(), resellerID, req.Count, req.Value)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Convert to DTOs
	out := make([]*dto.RedeemCode, 0, len(codes))
	for i := range codes {
		out = append(out, dto.RedeemCodeFromService(&codes[i]))
	}

	response.Success(c, out)
}

// List returns redeem codes owned by the reseller
// GET /api/v1/reseller/redeem
func (h *RedeemHandler) List(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	codes, pag, err := h.resellerService.ListRedeemCodes(c.Request.Context(), resellerID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.RedeemCode, 0, len(codes))
	for i := range codes {
		out = append(out, dto.RedeemCodeFromService(&codes[i]))
	}

	response.PaginatedWithResult(c, out, &response.PaginationResult{
		Total:    pag.Total,
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Pages:    pag.Pages,
	})
}

// Delete deletes a redeem code owned by the reseller
// DELETE /api/v1/reseller/redeem/:id
func (h *RedeemHandler) Delete(c *gin.Context) {
	resellerID := getResellerIDFromContext(c)
	if resellerID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	codeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid code ID")
		return
	}

	if err := h.resellerService.DeleteRedeemCode(c.Request.Context(), resellerID, codeID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, nil)
}
