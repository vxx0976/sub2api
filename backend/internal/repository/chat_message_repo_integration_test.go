//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/stretchr/testify/suite"
)

type ChatMessageRepoSuite struct {
	suite.Suite
	ctx    context.Context
	repo   service.ChatMessageRepository
	convID int64
}

func (s *ChatMessageRepoSuite) SetupTest() {
	s.ctx = context.Background()
	tx := testEntTx(s.T())
	s.repo = NewChatMessageRepository(tx.Client())

	convRepo := NewChatConversationRepository(tx.Client())
	conv := &service.ChatConversation{VisitorName: "访客", Status: service.ChatConversationStatusOpen}
	s.Require().NoError(convRepo.Create(s.ctx, conv))
	s.convID = conv.ID
}

func TestChatMessageRepoSuite(t *testing.T) {
	suite.Run(t, new(ChatMessageRepoSuite))
}

func (s *ChatMessageRepoSuite) seed(n int) {
	for i := 1; i <= n; i++ {
		msg := &service.ChatMessage{
			ConversationID: s.convID,
			SenderType:     service.ChatSenderTypeVisitor,
			Content:        fmt.Sprintf("msg-%03d", i),
		}
		s.Require().NoError(s.repo.Create(s.ctx, msg))
	}
}

// 负 offset = "最后一页":会话消息多于一页时,默认要拿最近的 limit 条(仍按时间正序)。
func (s *ChatMessageRepoSuite) TestListByConversation_NegativeOffsetReturnsLatestPage() {
	s.seed(120)

	msgs, total, err := s.repo.ListByConversation(s.ctx, s.convID, 50, -1)
	s.Require().NoError(err)
	s.Require().Equal(120, total)
	s.Require().Len(msgs, 50)
	s.Require().Equal("msg-071", msgs[0].Content)
	s.Require().Equal("msg-120", msgs[49].Content)
}

// 消息不足一页时,负 offset 不应把 offset 算成负数导致报错或丢消息。
func (s *ChatMessageRepoSuite) TestListByConversation_NegativeOffsetShorterThanPage() {
	s.seed(3)

	msgs, total, err := s.repo.ListByConversation(s.ctx, s.convID, 50, -1)
	s.Require().NoError(err)
	s.Require().Equal(3, total)
	s.Require().Len(msgs, 3)
	s.Require().Equal("msg-001", msgs[0].Content)
	s.Require().Equal("msg-003", msgs[2].Content)
}

// 显式 offset 仍是原来的从头分页语义,翻历史不受影响。
func (s *ChatMessageRepoSuite) TestListByConversation_ExplicitOffsetUnchanged() {
	s.seed(120)

	msgs, _, err := s.repo.ListByConversation(s.ctx, s.convID, 50, 0)
	s.Require().NoError(err)
	s.Require().Len(msgs, 50)
	s.Require().Equal("msg-001", msgs[0].Content)
	s.Require().Equal("msg-050", msgs[49].Content)
}
