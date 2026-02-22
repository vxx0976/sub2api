package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// OrderHandler handles admin order management
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler creates a new admin order handler
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// List handles listing all orders with pagination
// GET /api/v1/admin/orders
func (h *OrderHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")
	search := c.Query("search")

	orders, pag, err := h.orderService.GetAllOrders(c.Request.Context(), page, pageSize, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, orders, pag.Total, page, pageSize)
}

// Delete handles deleting a pending order
// DELETE /api/v1/admin/orders/:id
func (h *OrderHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	if err := h.orderService.DeleteOrder(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, nil)
}

// MarkPaid handles marking a pending order as paid manually
// POST /api/v1/admin/orders/:id/mark-paid
func (h *OrderHandler) MarkPaid(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	if err := h.orderService.AdminMarkOrderPaid(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, nil)
}
