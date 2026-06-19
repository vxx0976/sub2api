package reseller

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// ServiceTokenHandler 管理分销商的 M2M 服务令牌（签发 / 列表 / 吊销）。
// 这些接口走分销商 JWT 登录鉴权，由站长/分销商在后台操作。
type ServiceTokenHandler struct {
	tokenService *service.ResellerAPITokenService
}

// NewServiceTokenHandler 创建 ServiceTokenHandler。
func NewServiceTokenHandler(tokenService *service.ResellerAPITokenService) *ServiceTokenHandler {
	return &ServiceTokenHandler{tokenService: tokenService}
}

type createServiceTokenRequest struct {
	Name          string `json:"name"`
	ExpiresInDays *int   `json:"expires_in_days"` // nil 或 <=0 表示永不过期
}

// serviceTokenView 是返回给前端的安全视图（不含明文，也不含哈希）。
type serviceTokenView struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Status      string     `json:"status"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toServiceTokenView(t *service.ResellerAPIToken) serviceTokenView {
	return serviceTokenView{
		ID:          t.ID,
		Name:        t.Name,
		TokenPrefix: t.TokenPrefix,
		Status:      t.Status,
		LastUsedAt:  t.LastUsedAt,
		ExpiresAt:   t.ExpiresAt,
		CreatedAt:   t.CreatedAt,
	}
}

// Create 签发一个新的服务令牌，明文仅在本次响应中返回一次。
func (h *ServiceTokenHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createServiceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var ttl *time.Duration
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		d := time.Duration(*req.ExpiresInDays) * 24 * time.Hour
		ttl = &d
	}

	plaintext, token, err := h.tokenService.GenerateToken(c.Request.Context(), subject.UserID, req.Name, ttl)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	view := toServiceTokenView(token)
	response.Success(c, gin.H{
		"token": plaintext, // 仅此一次可见，请妥善保存
		"info":  view,
	})
}

// List 返回分销商的全部服务令牌（安全视图）。
func (h *ServiceTokenHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tokens, err := h.tokenService.ListTokens(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	views := make([]serviceTokenView, 0, len(tokens))
	for _, t := range tokens {
		views = append(views, toServiceTokenView(t))
	}
	response.Success(c, views)
}

// Revoke 吊销分销商自己的某个服务令牌。
func (h *ServiceTokenHandler) Revoke(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tokenID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid token ID")
		return
	}

	if err := h.tokenService.RevokeToken(c.Request.Context(), subject.UserID, tokenID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Token revoked"})
}
