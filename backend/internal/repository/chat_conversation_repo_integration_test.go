//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/stretchr/testify/suite"
)

type ChatConversationRepoSuite struct {
	suite.Suite
	ctx  context.Context
	repo service.ChatConversationRepository
}

func (s *ChatConversationRepoSuite) SetupTest() {
	s.ctx = context.Background()
	tx := testEntTx(s.T())
	s.repo = NewChatConversationRepository(tx.Client())
}

func TestChatConversationRepoSuite(t *testing.T) {
	suite.Run(t, new(ChatConversationRepoSuite))
}

// createConversation 建一个会话;lastMessageAt 为零值表示"点开气泡但一句话没聊"。
func (s *ChatConversationRepoSuite) createConversation(name string, lastMessageAt time.Time) int64 {
	conv := &service.ChatConversation{
		VisitorName: name,
		Status:      service.ChatConversationStatusOpen,
	}
	s.Require().NoError(s.repo.Create(s.ctx, conv))
	if !lastMessageAt.IsZero() {
		s.Require().NoError(s.repo.UpdateLastMessage(s.ctx, conv.ID, "hi", lastMessageAt))
	}
	return conv.ID
}

func (s *ChatConversationRepoSuite) TestList_SkipsEmptyAndSortsNewestFirst() {
	now := time.Now().UTC()
	s.createConversation("empty", time.Time{})
	s.createConversation("older", now.Add(-2*time.Hour))
	s.createConversation("newest", now.Add(-1*time.Minute))

	items, result, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, service.ChatConversationListFilters{})
	s.Require().NoError(err)
	s.Require().Len(items, 2)
	s.Require().Equal("newest", items[0].VisitorName)
	s.Require().Equal("older", items[1].VisitorName)
	s.Require().EqualValues(2, result.Total)
}

func (s *ChatConversationRepoSuite) TestList_EmptyConversationExcludedFromSearchAndStatusFilters() {
	s.createConversation("silent-visitor", time.Time{})

	items, result, err := s.repo.List(
		s.ctx,
		pagination.PaginationParams{Page: 1, PageSize: 10},
		service.ChatConversationListFilters{Status: service.ChatConversationStatusOpen, Search: "silent"},
	)
	s.Require().NoError(err)
	s.Require().Empty(items)
	s.Require().EqualValues(0, result.Total)
}
