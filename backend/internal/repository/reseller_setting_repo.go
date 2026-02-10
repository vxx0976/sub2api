package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/resellersetting"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type resellerSettingRepository struct {
	client *dbent.Client
}

func NewResellerSettingRepository(client *dbent.Client) service.ResellerSettingRepository {
	return &resellerSettingRepository{client: client}
}

func (r *resellerSettingRepository) GetAll(ctx context.Context, resellerID int64) (map[string]string, error) {
	settings, err := r.client.ResellerSetting.Query().
		Where(resellersetting.ResellerIDEQ(resellerID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (r *resellerSettingRepository) Get(ctx context.Context, resellerID int64, key string) (string, error) {
	s, err := r.client.ResellerSetting.Query().
		Where(
			resellersetting.ResellerIDEQ(resellerID),
			resellersetting.KeyEQ(key),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return s.Value, nil
}

func (r *resellerSettingRepository) Set(ctx context.Context, resellerID int64, key, value string) error {
	// Upsert: insert or update on conflict
	return r.client.ResellerSetting.Create().
		SetResellerID(resellerID).
		SetKey(key).
		SetValue(value).
		OnConflictColumns(resellersetting.FieldResellerID, resellersetting.FieldKey).
		UpdateValue().
		UpdateUpdatedAt().
		Exec(ctx)
}

func (r *resellerSettingRepository) SetAll(ctx context.Context, resellerID int64, settings map[string]string) error {
	for key, value := range settings {
		if err := r.Set(ctx, resellerID, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *resellerSettingRepository) ListByKey(ctx context.Context, key string) (map[int64]string, error) {
	settings, err := r.client.ResellerSetting.Query().
		Where(resellersetting.KeyEQ(key)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(settings))
	for _, s := range settings {
		result[s.ResellerID] = s.Value
	}
	return result, nil
}
