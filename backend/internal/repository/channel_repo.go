package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/channel"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type channelRepository struct {
	client *dbent.Client
}

// NewChannelRepository creates a new channel repository
func NewChannelRepository(client *dbent.Client) service.ChannelRepository {
	return &channelRepository{client: client}
}

func (r *channelRepository) Create(ctx context.Context, ch *service.Channel) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Channel.Create().
		SetName(ch.Name).
		SetStatus(ch.Status).
		SetBalanceMethod(ch.BalanceMethod).
		SetBalanceUnit(ch.BalanceUnit)

	if ch.Description != nil {
		builder.SetDescription(*ch.Description)
	}
	if ch.Platform != nil {
		builder.SetPlatform(*ch.Platform)
	}
	if ch.BalanceURL != nil {
		builder.SetBalanceURL(*ch.BalanceURL)
	}
	if ch.BalanceHeaders != nil {
		builder.SetBalanceHeaders(ch.BalanceHeaders)
	}
	if ch.BalanceBody != nil {
		builder.SetBalanceBody(*ch.BalanceBody)
	}
	if ch.BalancePath != nil {
		builder.SetBalancePath(*ch.BalancePath)
	}
	if ch.IconURL != nil {
		builder.SetIconURL(*ch.IconURL)
	}
	if ch.WebsiteURL != nil {
		builder.SetWebsiteURL(*ch.WebsiteURL)
	}

	created, err := builder.Save(ctx)
	if err == nil {
		applyChannelEntityToService(ch, created)
	}
	return translatePersistenceError(err, nil, nil)
}

func (r *channelRepository) GetByID(ctx context.Context, id int64) (*service.Channel, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Channel.Query().
		Where(channel.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrChannelNotFound, nil)
	}
	return channelEntityToService(m), nil
}

func (r *channelRepository) GetByName(ctx context.Context, name string) (*service.Channel, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Channel.Query().
		Where(channel.NameEQ(name)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrChannelNotFound, nil)
	}
	return channelEntityToService(m), nil
}

func (r *channelRepository) Update(ctx context.Context, ch *service.Channel) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Channel.UpdateOneID(ch.ID).
		SetName(ch.Name).
		SetStatus(ch.Status).
		SetBalanceMethod(ch.BalanceMethod).
		SetBalanceUnit(ch.BalanceUnit)

	if ch.Description != nil {
		builder.SetDescription(*ch.Description)
	} else {
		builder.ClearDescription()
	}
	if ch.Platform != nil {
		builder.SetPlatform(*ch.Platform)
	} else {
		builder.ClearPlatform()
	}
	if ch.BalanceURL != nil {
		builder.SetBalanceURL(*ch.BalanceURL)
	} else {
		builder.ClearBalanceURL()
	}
	if ch.BalanceHeaders != nil {
		builder.SetBalanceHeaders(ch.BalanceHeaders)
	} else {
		builder.ClearBalanceHeaders()
	}
	if ch.BalanceBody != nil {
		builder.SetBalanceBody(*ch.BalanceBody)
	} else {
		builder.ClearBalanceBody()
	}
	if ch.BalancePath != nil {
		builder.SetBalancePath(*ch.BalancePath)
	} else {
		builder.ClearBalancePath()
	}
	if ch.IconURL != nil {
		builder.SetIconURL(*ch.IconURL)
	} else {
		builder.ClearIconURL()
	}
	if ch.WebsiteURL != nil {
		builder.SetWebsiteURL(*ch.WebsiteURL)
	} else {
		builder.ClearWebsiteURL()
	}

	updated, err := builder.Save(ctx)
	if err == nil {
		applyChannelEntityToService(ch, updated)
	}
	return translatePersistenceError(err, service.ErrChannelNotFound, nil)
}

func (r *channelRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	err := client.Channel.DeleteOneID(id).Exec(ctx)
	return translatePersistenceError(err, service.ErrChannelNotFound, nil)
}

func (r *channelRepository) List(ctx context.Context, params pagination.PaginationParams, platform string, status string, search string) ([]service.Channel, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.Channel.Query()

	// 过滤平台
	if platform != "" {
		q = q.Where(channel.PlatformEQ(platform))
	}

	// 过滤状态
	if status != "" {
		q = q.Where(channel.StatusEQ(status))
	}

	// 搜索名称
	if search != "" {
		q = q.Where(channel.NameContains(search))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	channels, err := q.
		Order(dbent.Desc(channel.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return channelEntitiesToService(channels), paginationResultFromTotal(int64(total), params), nil
}

func (r *channelRepository) UpdateBalance(ctx context.Context, id int64, balance *float64, lastCheckAt time.Time, lastError *string) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Channel.UpdateOneID(id).
		SetLastCheckAt(lastCheckAt)

	if balance != nil {
		builder.SetCachedBalance(*balance)
		builder.ClearLastError()
		builder.SetStatus(service.ChannelStatusActive)
	} else {
		builder.ClearCachedBalance()
	}

	if lastError != nil {
		builder.SetLastError(*lastError)
		builder.SetStatus(service.ChannelStatusError)
	} else {
		builder.ClearLastError()
	}

	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrChannelNotFound, nil)
}

func channelEntityToService(m *dbent.Channel) *service.Channel {
	if m == nil {
		return nil
	}
	return &service.Channel{
		ID:             m.ID,
		Name:           m.Name,
		Description:    m.Description,
		Platform:       m.Platform,
		Status:         m.Status,
		IconURL:        m.IconURL,
		WebsiteURL:     m.WebsiteURL,
		BalanceURL:     m.BalanceURL,
		BalanceMethod:  m.BalanceMethod,
		BalanceHeaders: m.BalanceHeaders,
		BalanceBody:    m.BalanceBody,
		BalancePath:    m.BalancePath,
		BalanceUnit:    m.BalanceUnit,
		CachedBalance:  m.CachedBalance,
		LastCheckAt:    m.LastCheckAt,
		LastError:      m.LastError,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func channelEntitiesToService(models []*dbent.Channel) []service.Channel {
	out := make([]service.Channel, 0, len(models))
	for i := range models {
		if ch := channelEntityToService(models[i]); ch != nil {
			out = append(out, *ch)
		}
	}
	return out
}

func applyChannelEntityToService(dst *service.Channel, src *dbent.Channel) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}
