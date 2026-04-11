package repository

import (
	"context"
	"database/sql"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbfailover "github.com/Wei-Shaw/sub2api/ent/failovergroupevent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type failoverEventRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewFailoverEventRepository(client *dbent.Client, sqlDB *sql.DB) service.FailoverEventRepository {
	return &failoverEventRepository{client: client, sql: sqlDB}
}

func (r *failoverEventRepository) Create(ctx context.Context, event *service.FailoverGroupEvent) error {
	return createFailoverEventWithClient(ctx, r.client, event)
}

func (r *failoverEventRepository) ListByGroupID(ctx context.Context, virtualGroupID int64, limit int) ([]*service.FailoverGroupEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	events, err := r.client.FailoverGroupEvent.Query().
		Where(dbfailover.VirtualGroupIDEQ(virtualGroupID)).
		Order(dbent.Desc(dbfailover.FieldOccurredAt), dbent.Desc(dbfailover.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*service.FailoverGroupEvent, 0, len(events))
	for _, e := range events {
		out = append(out, failoverEventEntityToService(e))
	}
	return out, nil
}

func (r *failoverEventRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res, err := r.sql.ExecContext(ctx, `DELETE FROM failover_group_events WHERE occurred_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func failoverEventEntityToService(e *dbent.FailoverGroupEvent) *service.FailoverGroupEvent {
	if e == nil {
		return nil
	}
	return &service.FailoverGroupEvent{
		ID:             e.ID,
		VirtualGroupID: e.VirtualGroupID,
		FromMemberID:   e.FromMemberID,
		ToMemberID:     e.ToMemberID,
		Reason:         e.Reason,
		TriggeredBy:    e.TriggeredBy,
		Note:           e.Note,
		OccurredAt:     e.OccurredAt,
	}
}
