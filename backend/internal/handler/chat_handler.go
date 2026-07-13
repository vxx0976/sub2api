package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	chatService *service.ChatService
	chatHub     *service.ChatHub
}

func NewChatHandler(chatService *service.ChatService, chatHub *service.ChatHub) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		chatHub:     chatHub,
	}
}

var uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type createConversationRequest struct {
	GuestToken string `json:"guest_token"`
}

type sendMessageRequest struct {
	Content    string `json:"content" binding:"required"`
	GuestToken string `json:"guest_token"`
}

// CreateOrGetConversation creates or retrieves an existing open conversation.
// POST /api/v1/chat/conversations
func (h *ChatHandler) CreateOrGetConversation(c *gin.Context) {
	var req createConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	var guestToken *string
	var userID *int64
	visitorName := "访客"

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if ok && subject.UserID > 0 {
		userID = &subject.UserID
		visitorName = "用户"
	} else if req.GuestToken != "" {
		if !uuidRegexp.MatchString(req.GuestToken) {
			response.ErrorFrom(c, service.ErrChatInvalidGuestToken)
			return
		}
		guestToken = &req.GuestToken
	} else {
		response.BadRequest(c, "guest_token is required for anonymous visitors")
		return
	}

	conv, err := h.chatService.GetOrCreateConversation(c.Request.Context(), guestToken, userID, visitorName)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	msgs, _, _ := h.chatService.GetMessages(c.Request.Context(), conv.ID, 50, 0)
	response.Success(c, gin.H{
		"conversation": chatConversationDTO(conv),
		"messages":     chatMessagesDTO(msgs),
		"admin_online": h.chatHub.AdminCount() > 0,
	})
}

// GetOpenConversation returns the visitor's existing open conversation WITHOUT creating one.
// Used by the widget on page load to detect unread admin replies (red dot) without side effects.
// GET /api/v1/chat/conversation
func (h *ChatHandler) GetOpenConversation(c *gin.Context) {
	var guestToken *string
	var userID *int64

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if ok && subject.UserID > 0 {
		userID = &subject.UserID
	} else if gt := c.Query("guest_token"); gt != "" {
		if !uuidRegexp.MatchString(gt) {
			response.ErrorFrom(c, service.ErrChatInvalidGuestToken)
			return
		}
		guestToken = &gt
	}

	empty := gin.H{"conversation": nil, "messages": []gin.H{}, "admin_online": h.chatHub.AdminCount() > 0}
	if userID == nil && guestToken == nil {
		response.Success(c, empty)
		return
	}

	conv, err := h.chatService.GetOpenConversation(c.Request.Context(), guestToken, userID)
	if err != nil || conv == nil {
		response.Success(c, empty)
		return
	}

	msgs, _, _ := h.chatService.GetMessages(c.Request.Context(), conv.ID, 50, 0)
	response.Success(c, gin.H{
		"conversation": chatConversationDTO(conv),
		"messages":     chatMessagesDTO(msgs),
		"admin_online": h.chatHub.AdminCount() > 0,
	})
}

// SendMessage sends a message from a visitor.
// POST /api/v1/chat/conversations/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || convID <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	conv, err := h.chatService.GetConversation(c.Request.Context(), convID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if ok && subject.UserID > 0 {
		if conv.UserID == nil || *conv.UserID != subject.UserID {
			response.ErrorFrom(c, service.ErrChatUnauthorized)
			return
		}
	} else {
		if req.GuestToken == "" || !uuidRegexp.MatchString(req.GuestToken) {
			response.ErrorFrom(c, service.ErrChatInvalidGuestToken)
			return
		}
		if conv.GuestToken == nil || *conv.GuestToken != req.GuestToken {
			response.ErrorFrom(c, service.ErrChatUnauthorized)
			return
		}
	}

	msg, err := h.chatService.SendMessage(c.Request.Context(), convID, service.ChatSenderTypeVisitor, req.Content)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.chatHub.BroadcastToAdmins(service.ChatWSMessage{
		Type: "new_message",
		Data: gin.H{
			"conversation_id": convID,
			"message":         chatMessageDTO(msg),
		},
	})

	response.Success(c, chatMessageDTO(msg))
}

// GetMessages returns message history for a conversation.
// GET /api/v1/chat/conversations/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || convID <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	conv, err := h.chatService.GetConversation(c.Request.Context(), convID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if ok && subject.UserID > 0 {
		if conv.UserID == nil || *conv.UserID != subject.UserID {
			response.ErrorFrom(c, service.ErrChatUnauthorized)
			return
		}
	} else {
		guestToken := c.Query("guest_token")
		if guestToken == "" || !uuidRegexp.MatchString(guestToken) {
			response.ErrorFrom(c, service.ErrChatInvalidGuestToken)
			return
		}
		if conv.GuestToken == nil || *conv.GuestToken != guestToken {
			response.ErrorFrom(c, service.ErrChatUnauthorized)
			return
		}
	}

	limit := 50
	offset := 0
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		offset = v
	}

	msgs, total, err := h.chatService.GetMessages(c.Request.Context(), convID, limit, offset)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"messages": chatMessagesDTO(msgs),
		"total":    total,
	})
}

var visitorUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"sub2api-chat"},
}

// VisitorWebSocket handles WebSocket connections for visitors.
// GET /api/v1/chat/ws?conversation_id=X&guest_token=Y
// Logged-in browser clients carry their JWT in the jwt.* WebSocket subprotocol.
func (h *ChatHandler) VisitorWebSocket(c *gin.Context) {
	convIDStr := c.Query("conversation_id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil || convID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation_id"})
		return
	}

	conv, err := h.chatService.GetConversation(c.Request.Context(), convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if ok && subject.UserID > 0 {
		if conv.UserID == nil || *conv.UserID != subject.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		}
	} else {
		guestToken := c.Query("guest_token")
		if guestToken == "" || !uuidRegexp.MatchString(guestToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guest_token"})
			return
		}
		if conv.GuestToken == nil || *conv.GuestToken != guestToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		}
	}

	ws, err := visitorUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	conn := newGorillaChatWSConn(ws)
	h.chatHub.RegisterVisitor(convID, conn)
	defer func() {
		h.chatHub.UnregisterVisitor(convID)
		_ = ws.Close()
	}()

	conn.startWriter()
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

type gorillaChatWSConn struct {
	conn   *websocket.Conn
	sendCh chan []byte
	once   sync.Once
}

func newGorillaChatWSConn(conn *websocket.Conn) *gorillaChatWSConn {
	return &gorillaChatWSConn{
		conn:   conn,
		sendCh: make(chan []byte, 64),
	}
}

func (c *gorillaChatWSConn) Send(data []byte) error {
	select {
	case c.sendCh <- data:
		return nil
	default:
		return nil
	}
}

func (c *gorillaChatWSConn) Close() {
	c.once.Do(func() {
		close(c.sendCh)
	})
}

func (c *gorillaChatWSConn) startWriter() {
	go func() {
		for data := range c.sendCh {
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()
}

func chatConversationDTO(c *service.ChatConversation) gin.H {
	if c == nil {
		return nil
	}
	dto := gin.H{
		"id":                   c.ID,
		"visitor_name":         c.VisitorName,
		"status":               c.Status,
		"admin_unread_count":   c.AdminUnreadCount,
		"last_message_preview": c.LastMessagePreview,
		"created_at":           c.CreatedAt,
		"updated_at":           c.UpdatedAt,
	}
	if c.UserID != nil {
		dto["user_id"] = *c.UserID
	}
	if c.LastMessageAt != nil {
		dto["last_message_at"] = *c.LastMessageAt
	}
	return dto
}

func chatMessageDTO(m *service.ChatMessage) gin.H {
	if m == nil {
		return nil
	}
	return gin.H{
		"id":              m.ID,
		"conversation_id": m.ConversationID,
		"sender_type":     m.SenderType,
		"content":         m.Content,
		"created_at":      m.CreatedAt,
	}
}

func chatMessagesDTO(msgs []service.ChatMessage) []gin.H {
	out := make([]gin.H, 0, len(msgs))
	for i := range msgs {
		out = append(out, chatMessageDTO(&msgs[i]))
	}
	return out
}
