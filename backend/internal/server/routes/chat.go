package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterChatRoutes registers public chat routes (no mandatory auth).
// Optional JWT auth is applied so logged-in users can be identified.
func RegisterChatRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
) {
	chat := v1.Group("/chat")
	chat.Use(optionalJWTAuth(jwtAuth))
	{
		chat.POST("/conversations", h.Chat.CreateOrGetConversation)
		chat.GET("/conversation", h.Chat.GetOpenConversation)
		chat.GET("/conversations/:id/messages", h.Chat.GetMessages)
		chat.POST("/conversations/:id/messages", h.Chat.SendMessage)
		chat.GET("/ws", h.Chat.VisitorWebSocket)
	}
}

// optionalJWTAuth applies JWT auth only when an Authorization header is present.
// If valid, the user subject is set in context. If absent, the request proceeds anonymously.
func optionalJWTAuth(jwtAuth middleware.JWTAuthMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			gin.HandlerFunc(jwtAuth)(c)
			return
		}
		c.Next()
	}
}
