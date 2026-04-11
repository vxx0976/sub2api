package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type failoverGroupRepoStub struct {
	groupRepoNoop

	counts          map[int64]int
	transitionCalls []failoverTransitionCall
}

type failoverTransitionCall struct {
	groupID      int64
	toMemberID   int64
	reason       string
	fromMemberID *int64
}

func (s *failoverGroupRepoStub) CountSchedulableAccountsByGroup(_ context.Context, groupID int64) (int, error) {
	return s.counts[groupID], nil
}

func (s *failoverGroupRepoStub) TransitionFailoverActive(_ context.Context, groupID int64, newMemberID int64, expectedVersion int64, event *FailoverGroupEvent) (bool, error) {
	if expectedVersion < 0 {
		return false, nil
	}
	call := failoverTransitionCall{
		groupID:    groupID,
		toMemberID: newMemberID,
	}
	if event != nil {
		call.reason = event.Reason
		call.fromMemberID = event.FromMemberID
	}
	s.transitionCalls = append(s.transitionCalls, call)
	return true, nil
}

func TestFailoverGroupServiceSoftReconcileDemotesOnly(t *testing.T) {
	repo := &failoverGroupRepoStub{
		counts: map[int64]int{
			10: 1, // higher priority member A recovered
			20: 0, // current active member B unavailable
			30: 1, // lower priority member C available
		},
	}
	svc := NewFailoverGroupService(repo, nil, nil, nil, nil)
	active := int64(20)
	group := &Group{
		ID:                     99,
		IsFailoverGroup:        true,
		FailoverMemberIDs:      []int64{10, 20, 30},
		FailoverActiveMemberID: &active,
		FailoverActiveVersion:  7,
	}

	err := svc.softReconcileOne(context.Background(), group)
	require.NoError(t, err)
	require.Len(t, repo.transitionCalls, 1)
	require.Equal(t, int64(30), repo.transitionCalls[0].toMemberID, "soft reconcile should only demote to a later member")
	require.Equal(t, FailoverEventReasonSoftDemote, repo.transitionCalls[0].reason)
}
