package admin

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
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

// ListConversations returns a paginated list of conversations.
// GET /api/v1/admin/chat/conversations
func (h *ChatHandler) ListConversations(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))
	sortBy := c.DefaultQuery("sort_by", "last_message_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	items, paginationResult, err := h.chatService.ListConversations(
		c.Request.Context(),
		params,
		service.ChatConversationListFilters{Status: status, Search: search},
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]gin.H, 0, len(items))
	nameCache := make(map[int64]string)
	for i := range items {
		name := h.resolveDisplayName(c.Request.Context(), &items[i], nameCache)
		out = append(out, adminChatConversationDTO(&items[i], name))
	}
	response.Paginated(c, out, paginationResult.Total, page, pageSize)
}

// GetConversation returns a single conversation.
// GET /api/v1/admin/chat/conversations/:id
func (h *ChatHandler) GetConversation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	conv, err := h.chatService.GetConversation(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, adminChatConversationDTO(conv, h.resolveDisplayName(c.Request.Context(), conv, nil)))
}

// GetMessages returns messages for a conversation.
// GET /api/v1/admin/chat/conversations/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || convID <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	limit := 50
	// 不传 offset 时取最后一页,即最近的 limit 条消息
	offset := -1
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

	out := make([]gin.H, 0, len(msgs))
	for i := range msgs {
		out = append(out, adminChatMessageDTO(&msgs[i]))
	}
	response.Success(c, gin.H{
		"messages": out,
		"total":    total,
	})
}

type adminSendReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

// SendReply sends a reply from admin.
// POST /api/v1/admin/chat/conversations/:id/messages
func (h *ChatHandler) SendReply(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || convID <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	var req adminSendReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	msg, err := h.chatService.SendMessage(c.Request.Context(), convID, service.ChatSenderTypeAdmin, req.Content)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.chatHub.SendToVisitor(convID, service.ChatWSMessage{
		Type: "new_message",
		Data: gin.H{
			"message": adminChatMessageDTO(msg),
		},
	})

	h.chatHub.BroadcastToAdmins(service.ChatWSMessage{
		Type: "admin_reply",
		Data: gin.H{
			"conversation_id": convID,
			"message":         adminChatMessageDTO(msg),
		},
	})

	response.Success(c, adminChatMessageDTO(msg))
}

// CloseConversation closes a conversation.
// POST /api/v1/admin/chat/conversations/:id/close
func (h *ChatHandler) CloseConversation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	if err := h.chatService.CloseConversation(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	h.chatHub.SendToVisitor(id, service.ChatWSMessage{
		Type: "conversation_closed",
		Data: gin.H{"conversation_id": id},
	})

	response.Success(c, gin.H{"message": "ok"})
}

// MarkRead resets the unread counter for a conversation.
// POST /api/v1/admin/chat/conversations/:id/read
func (h *ChatHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid conversation ID")
		return
	}

	if err := h.chatService.MarkAdminRead(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

// GetUnreadCount returns total unread conversation count.
// GET /api/v1/admin/chat/unread-count
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	count, err := h.chatService.CountUnread(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

var adminChatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"sub2api-admin"},
}

// AdminWebSocket handles WebSocket connections for admin chat.
// GET /api/v1/admin/chat/ws
func (h *ChatHandler) AdminWebSocket(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ws, err := adminChatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	conn := newAdminChatWSConn(ws)
	h.chatHub.RegisterAdmin(subject.UserID, conn)
	defer func() {
		h.chatHub.UnregisterAdmin(subject.UserID)
		_ = ws.Close()
	}()

	conn.startWriter()

	ws.SetPingHandler(func(appData string) error {
		return ws.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
	})

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

type adminChatWSConn struct {
	conn   *websocket.Conn
	sendCh chan []byte
	once   sync.Once
}

func newAdminChatWSConn(conn *websocket.Conn) *adminChatWSConn {
	return &adminChatWSConn{
		conn:   conn,
		sendCh: make(chan []byte, 64),
	}
}

func (c *adminChatWSConn) Send(data []byte) error {
	select {
	case c.sendCh <- data:
		return nil
	default:
		return nil
	}
}

func (c *adminChatWSConn) Close() {
	c.once.Do(func() {
		close(c.sendCh)
	})
}

func (c *adminChatWSConn) startWriter() {
	go func() {
		for data := range c.sendCh {
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()
}

// resolveDisplayName 解析管理员列表展示名(登录用户名/邮箱, 访客显示"访客"),
// 按 user_id 在本次请求内缓存去重以避免 N+1 查询。cache 传 nil 表示不缓存。
func (h *ChatHandler) resolveDisplayName(ctx context.Context, conv *service.ChatConversation, cache map[int64]string) string {
	if conv == nil || conv.UserID == nil {
		return h.chatService.ResolveDisplayName(ctx, nil)
	}
	if cache != nil {
		if name, ok := cache[*conv.UserID]; ok {
			return name
		}
	}
	name := h.chatService.ResolveDisplayName(ctx, conv.UserID)
	if cache != nil {
		cache[*conv.UserID] = name
	}
	return name
}

func adminChatConversationDTO(c *service.ChatConversation, displayName string) gin.H {
	if c == nil {
		return nil
	}
	dto := gin.H{
		"id":                   c.ID,
		"visitor_name":         c.VisitorName,
		"display_name":         displayName,
		"status":               c.Status,
		"admin_unread_count":   c.AdminUnreadCount,
		"last_message_preview": c.LastMessagePreview,
		"created_at":           c.CreatedAt,
		"updated_at":           c.UpdatedAt,
	}
	if c.GuestToken != nil {
		dto["guest_token"] = *c.GuestToken
	}
	if c.UserID != nil {
		dto["user_id"] = *c.UserID
	}
	if c.LastMessageAt != nil {
		dto["last_message_at"] = *c.LastMessageAt
	}
	return dto
}

func adminChatMessageDTO(m *service.ChatMessage) gin.H {
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
