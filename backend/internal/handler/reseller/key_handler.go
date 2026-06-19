package reseller

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// resellerKeyCreateIdempotencyTTL 控制带 Idempotency-Key 的建 key 幂等记录留存时长。
// 取较短窗口：重试通常在首次失败后数分钟内发生，而回放体包含新 key 明文，需尽量缩短其留存。
const resellerKeyCreateIdempotencyTTL = 15 * time.Minute

// KeyHandler handles reseller API key management
type KeyHandler struct {
	resellerService *service.ResellerService
}

// NewKeyHandler creates a new KeyHandler
func NewKeyHandler(resellerService *service.ResellerService) *KeyHandler {
	return &KeyHandler{resellerService: resellerService}
}

// List returns the reseller's API keys
func (h *KeyHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	keys, pag, err := h.resellerService.ListKeys(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Convert service models to DTOs for proper JSON serialization
	dtoKeys := make([]*dto.APIKey, 0, len(keys))
	for i := range keys {
		dtoKeys = append(dtoKeys, dto.APIKeyFromService(&keys[i]))
	}

	response.PaginatedWithResult(c, dtoKeys, &response.PaginationResult{
		Total:    pag.Total,
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Pages:    pag.Pages,
	})
}

// Create creates a new API key for the reseller
func (h *KeyHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input service.CreateResellerKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	execute := func(ctx context.Context) (any, error) {
		key, err := h.resellerService.CreateKey(ctx, subject.UserID, &input)
		if err != nil {
			return nil, err
		}
		return dto.APIKeyFromService(key), nil
	}

	// 可选幂等：仅当请求带 Idempotency-Key 头时去重，避免 M2M 后端重试导致重复建 key；
	// 不带该头（如后台手动创建）时维持原有行为。
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	coordinator := service.DefaultIdempotencyCoordinator()
	if idempotencyKey == "" || coordinator == nil {
		data, err := execute(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, data)
		return
	}

	// Scope 折入 reseller id：令每个分销商拥有独立的幂等命名空间，
	// 避免两个分销商各自选用相同 Idempotency-Key 时发生跨租户 409 冲突。
	actorScope := "user:" + strconv.FormatInt(subject.UserID, 10)
	result, err := coordinator.Execute(c.Request.Context(), service.IdempotencyExecuteOptions{
		Scope:          "reseller_key_create:" + actorScope,
		ActorScope:     actorScope,
		Method:         c.Request.Method,
		Route:          c.FullPath(),
		IdempotencyKey: idempotencyKey,
		Payload:        input,
		RequireKey:     true,
		// 短 TTL：幂等重试只发生在首次请求失败后的数分钟内；
		// 回放需返回可用的新 key 明文，故缩短其在 idempotency_records 中的留存窗口，
		// 降低明文密钥在该表的二次副本暴露时间（api_keys 本就明文存储，此处仅控制额外副本）。
		TTL: resellerKeyCreateIdempotencyTTL,
	}, execute)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if result.Replayed {
		c.Header("X-Idempotency-Replayed", "true")
	}
	response.Success(c, result.Data)
}

// Update updates an API key owned by the reseller
func (h *KeyHandler) Update(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	var input service.UpdateResellerKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	key, err := h.resellerService.UpdateKey(c.Request.Context(), subject.UserID, keyID, &input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.APIKeyFromService(key))
}

// Delete deletes an API key owned by the reseller
func (h *KeyHandler) Delete(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	if err := h.resellerService.DeleteKey(c.Request.Context(), subject.UserID, keyID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Key deleted"})
}

// Enable activates an API key owned by the reseller.
func (h *KeyHandler) Enable(c *gin.Context) {
	h.setStatus(c, "active")
}

// Disable deactivates an API key owned by the reseller.
func (h *KeyHandler) Disable(c *gin.Context) {
	h.setStatus(c, "disabled")
}

// setStatus is a shared helper for Enable/Disable.
func (h *KeyHandler) setStatus(c *gin.Context, status string) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	key, err := h.resellerService.SetKeyStatus(c.Request.Context(), subject.UserID, keyID, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.APIKeyFromService(key))
}

// ResetQuota resets the quota of an API key
func (h *KeyHandler) ResetQuota(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid key ID")
		return
	}

	key, err := h.resellerService.ResetKeyQuota(c.Request.Context(), subject.UserID, keyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.APIKeyFromService(key))
}
