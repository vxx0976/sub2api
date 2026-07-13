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

// optionalJWTAuth applies JWT auth when an Authorization header or browser
// WebSocket jwt.* subprotocol token is present. Otherwise the request proceeds anonymously.
func optionalJWTAuth(jwtAuth middleware.JWTAuthMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" && middleware.IsWebSocketUpgradeRequest(c) {
			if token := middleware.ExtractJWTFromWebSocketSubprotocol(c); token != "" {
				c.Request.Header.Set("Authorization", "Bearer "+token)
			}
		}
		if c.GetHeader("Authorization") != "" {
			gin.HandlerFunc(jwtAuth)(c)
			return
		}
		c.Next()
	}
}
