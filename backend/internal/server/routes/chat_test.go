package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOptionalJWTAuthUsesWebSocketSubprotocolToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authCalled := false
	auth := servermiddleware.JWTAuthMiddleware(func(c *gin.Context) {
		authCalled = true
		require.Equal(t, "Bearer header.payload.signature", c.GetHeader("Authorization"))
		c.Next()
	})

	router := gin.New()
	router.GET("/chat/ws", optionalJWTAuth(auth), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/chat/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Protocol", "sub2api-chat, jwt.header.payload.signature")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNoContent, resp.Code)
	require.True(t, authCalled)
}

func TestOptionalJWTAuthDoesNotTrustSubprotocolOutsideWebSocketUpgrade(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authCalled := false
	auth := servermiddleware.JWTAuthMiddleware(func(c *gin.Context) {
		authCalled = true
		c.Next()
	})

	router := gin.New()
	router.GET("/chat", optionalJWTAuth(auth), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/chat", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "sub2api-chat, jwt.header.payload.signature")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusNoContent, resp.Code)
	require.False(t, authCalled)
}
