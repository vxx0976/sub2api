package admin

import (
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
