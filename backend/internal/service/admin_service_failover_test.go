package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type adminFailoverGroupRepoStub struct {
	groupRepoNoop

	groups map[int64]*Group

	updatedGroup                 *Group
	updateFailoverConfigGroupID  int64
	updateFailoverConfigMembers  []int64
	updateFailoverConfigActiveID *int64

	updatePinStateCalls     int
	updatePinStateGroupID   int64
	updatePinStateMemberID  *int64
	updatePinStateExpiresAt *time.Time
	updatePinStateEvent     *FailoverGroupEvent
}

func (s *adminFailoverGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	group := s.groups[id]
	if group == nil {
		return nil, ErrGroupNotFound
	}
	return cloneFailoverTestGroup(group), nil
}

func (s *adminFailoverGroupRepoStub) GetByIDLite(_ context.Context, id int64) (*Group, error) {
	group := s.groups[id]
	if group == nil {
		return nil, ErrGroupNotFound
	}
	return cloneFailoverTestGroup(group), nil
}

func (s *adminFailoverGroupRepoStub) Update(_ context.Context, group *Group) error {
	s.updatedGroup = cloneFailoverTestGroup(group)
	s.groups[group.ID] = cloneFailoverTestGroup(group)
	return nil
}

func (s *adminFailoverGroupRepoStub) UpdateFailoverConfig(_ context.Context, groupID int64, _ bool, memberIDs []int64, activeMemberID *int64) error {
	s.updateFailoverConfigGroupID = groupID
	s.updateFailoverConfigMembers = append([]int64(nil), memberIDs...)
	s.updateFailoverConfigActiveID = cloneFailoverTestInt64Ptr(activeMemberID)

	group := cloneFailoverTestGroup(s.groups[groupID])
	group.FailoverMemberIDs = append([]int64(nil), memberIDs...)
	group.FailoverActiveMemberID = cloneFailoverTestInt64Ptr(activeMemberID)
	s.groups[groupID] = group
	return nil
}

func (s *adminFailoverGroupRepoStub) UpdateFailoverPinState(_ context.Context, groupID int64, memberID *int64, expiresAt *time.Time, event *FailoverGroupEvent) error {
	s.updatePinStateCalls++
	s.updatePinStateGroupID = groupID
	s.updatePinStateMemberID = cloneFailoverTestInt64Ptr(memberID)
	s.updatePinStateExpiresAt = cloneFailoverTestTimePtr(expiresAt)
	s.updatePinStateEvent = cloneFailoverTestEvent(event)

	group := cloneFailoverTestGroup(s.groups[groupID])
	group.FailoverPinMemberID = cloneFailoverTestInt64Ptr(memberID)
	group.FailoverPinExpiresAt = cloneFailoverTestTimePtr(expiresAt)
	s.groups[groupID] = group
	return nil
}

func TestAdminServiceUpdateGroupClearsPinWhenPinnedMemberRemoved(t *testing.T) {
	pinExpiresAt := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	repo := &adminFailoverGroupRepoStub{
		groups: map[int64]*Group{
			100: {
				ID:                     100,
				Name:                   "smart-router",
				Platform:               PlatformAnthropic,
				Status:                 StatusActive,
				IsFailoverGroup:        true,
				FailoverMemberIDs:      []int64{1, 2, 3},
				FailoverActiveMemberID: cloneFailoverTestInt64(3),
				FailoverPinMemberID:    cloneFailoverTestInt64(3),
				FailoverPinExpiresAt:   &pinExpiresAt,
			},
			1: {ID: 1, Name: "group-a", Platform: PlatformAnthropic, Status: StatusActive},
			2: {ID: 2, Name: "group-b", Platform: PlatformAnthropic, Status: StatusActive},
			3: {ID: 3, Name: "group-c", Platform: PlatformAnthropic, Status: StatusActive},
		},
	}
	svc := &adminServiceImpl{groupRepo: repo}

	newMembers := []int64{1, 2}
	group, err := svc.UpdateGroup(context.Background(), 100, &UpdateGroupInput{
		FailoverMemberIDs: &newMembers,
	})

	require.NoError(t, err)
	require.Equal(t, []int64{1, 2}, group.FailoverMemberIDs)
	require.NotNil(t, group.FailoverActiveMemberID)
	require.Equal(t, int64(1), *group.FailoverActiveMemberID)
	require.Nil(t, group.FailoverPinMemberID)
	require.Nil(t, group.FailoverPinExpiresAt)

	require.Equal(t, int64(100), repo.updateFailoverConfigGroupID)
	require.Equal(t, []int64{1, 2}, repo.updateFailoverConfigMembers)
	require.NotNil(t, repo.updateFailoverConfigActiveID)
	require.Equal(t, int64(1), *repo.updateFailoverConfigActiveID)

	require.Equal(t, 1, repo.updatePinStateCalls)
	require.Equal(t, int64(100), repo.updatePinStateGroupID)
	require.Nil(t, repo.updatePinStateMemberID)
	require.Nil(t, repo.updatePinStateExpiresAt)
	require.NotNil(t, repo.updatePinStateEvent)
	require.Equal(t, FailoverEventReasonManualUnpin, repo.updatePinStateEvent.Reason)
	require.NotNil(t, repo.updatePinStateEvent.Note)
	require.Contains(t, *repo.updatePinStateEvent.Note, "pin target removed from member list")
	require.NotNil(t, repo.updatePinStateEvent.FromMemberID)
	require.Equal(t, int64(1), *repo.updatePinStateEvent.FromMemberID)
	require.NotNil(t, repo.updatePinStateEvent.ToMemberID)
	require.Equal(t, int64(1), *repo.updatePinStateEvent.ToMemberID)
}

func TestAdminServiceCreateAccountRejectsFailoverGroupBinding(t *testing.T) {
	repo := &adminFailoverGroupRepoStub{
		groups: map[int64]*Group{
			77: {
				ID:              77,
				Name:            "smart-router",
				Platform:        PlatformAnthropic,
				Status:          StatusActive,
				IsFailoverGroup: true,
			},
		},
	}
	svc := &adminServiceImpl{groupRepo: repo}

	account, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                  "acc-1",
		Platform:              PlatformAnthropic,
		Type:                  AccountTypeAPIKey,
		GroupIDs:              []int64{77},
		SkipMixedChannelCheck: true,
	})

	require.Nil(t, account)
	require.Error(t, err)
	require.Contains(t, err.Error(), "group 77 is a failover group and cannot bind accounts")
}

func cloneFailoverTestGroup(group *Group) *Group {
	if group == nil {
		return nil
	}
	cp := *group
	cp.FailoverMemberIDs = append([]int64(nil), group.FailoverMemberIDs...)
	cp.FailoverActiveMemberID = cloneFailoverTestInt64Ptr(group.FailoverActiveMemberID)
	cp.FailoverPinMemberID = cloneFailoverTestInt64Ptr(group.FailoverPinMemberID)
	cp.FailoverPinExpiresAt = cloneFailoverTestTimePtr(group.FailoverPinExpiresAt)
	return &cp
}

func cloneFailoverTestEvent(event *FailoverGroupEvent) *FailoverGroupEvent {
	if event == nil {
		return nil
	}
	cp := *event
	cp.FromMemberID = cloneFailoverTestInt64Ptr(event.FromMemberID)
	cp.ToMemberID = cloneFailoverTestInt64Ptr(event.ToMemberID)
	if event.Note != nil {
		note := *event.Note
		cp.Note = &note
	}
	if event.TriggeredBy != nil {
		triggeredBy := *event.TriggeredBy
		cp.TriggeredBy = &triggeredBy
	}
	return &cp
}

func cloneFailoverTestInt64(v int64) *int64 {
	return &v
}

func cloneFailoverTestInt64Ptr(v *int64) *int64 {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneFailoverTestTimePtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}
