package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/resellerapitoken"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type resellerAPITokenRepository struct {
	client *dbent.Client
}

// NewResellerAPITokenRepository 创建分销商服务令牌仓储。
func NewResellerAPITokenRepository(client *dbent.Client) service.ResellerAPITokenRepository {
	return &resellerAPITokenRepository{client: client}
}

func (r *resellerAPITokenRepository) Create(ctx context.Context, token *service.ResellerAPIToken) (*service.ResellerAPIToken, error) {
	create := r.client.ResellerAPIToken.Create().
		SetResellerID(token.ResellerID).
		SetName(token.Name).
		SetTokenPrefix(token.TokenPrefix).
		SetTokenHash(token.TokenHash).
		SetStatus(token.Status).
		SetNillableExpiresAt(token.ExpiresAt)

	row, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return entResellerAPITokenToModel(row), nil
}

func (r *resellerAPITokenRepository) GetByHash(ctx context.Context, hash string) (*service.ResellerAPIToken, error) {
	row, err := r.client.ResellerAPIToken.Query().
		Where(resellerapitoken.TokenHashEQ(hash)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entResellerAPITokenToModel(row), nil
}

func (r *resellerAPITokenRepository) GetByIDForReseller(ctx context.Context, id, resellerID int64) (*service.ResellerAPIToken, error) {
	row, err := r.client.ResellerAPIToken.Query().
		Where(
			resellerapitoken.IDEQ(id),
			resellerapitoken.ResellerIDEQ(resellerID),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entResellerAPITokenToModel(row), nil
}

func (r *resellerAPITokenRepository) ListByReseller(ctx context.Context, resellerID int64) ([]*service.ResellerAPIToken, error) {
	rows, err := r.client.ResellerAPIToken.Query().
		Where(resellerapitoken.ResellerIDEQ(resellerID)).
		Order(dbent.Desc(resellerapitoken.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*service.ResellerAPIToken, 0, len(rows))
	for _, row := range rows {
		out = append(out, entResellerAPITokenToModel(row))
	}
	return out, nil
}

func (r *resellerAPITokenRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.client.ResellerAPIToken.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
}

func (r *resellerAPITokenRepository) TouchLastUsed(ctx context.Context, id int64, at time.Time) error {
	return r.client.ResellerAPIToken.UpdateOneID(id).
		SetLastUsedAt(at).
		Exec(ctx)
}

func entResellerAPITokenToModel(row *dbent.ResellerAPIToken) *service.ResellerAPIToken {
	if row == nil {
		return nil
	}
	return &service.ResellerAPIToken{
		ID:          row.ID,
		ResellerID:  row.ResellerID,
		Name:        row.Name,
		TokenPrefix: row.TokenPrefix,
		TokenHash:   row.TokenHash,
		Status:      row.Status,
		LastUsedAt:  row.LastUsedAt,
		ExpiresAt:   row.ExpiresAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
