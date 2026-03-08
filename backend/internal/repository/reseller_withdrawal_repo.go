package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/resellerwithdrawal"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ResellerWithdrawalRepo struct {
	client *dbent.Client
}

func NewResellerWithdrawalRepo(client *dbent.Client) *ResellerWithdrawalRepo {
	return &ResellerWithdrawalRepo{client: client}
}

func (r *ResellerWithdrawalRepo) clientFromCtx(ctx context.Context) *dbent.Client {
	return clientFromContext(ctx, r.client)
}

// CreateIfNoPending 在事务中原子性地检查是否存在 pending 提现，若无则创建。
// 防止并发请求通过时间窗口绕过"只允许一个待审核提现"的约束。
func (r *ResellerWithdrawalRepo) CreateIfNoPending(ctx context.Context, w *service.ResellerWithdrawal) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}

	exists, err := tx.ResellerWithdrawal.Query().
		Where(resellerwithdrawal.ResellerIDEQ(w.ResellerID), resellerwithdrawal.StatusEQ("pending")).
		Exist(ctx)
	if err != nil {
		_ = tx.Rollback()
		return translatePersistenceError(err, nil, nil)
	}
	if exists {
		_ = tx.Rollback()
		return service.ErrWithdrawalPending
	}

	record, err := tx.ResellerWithdrawal.Create().
		SetResellerID(w.ResellerID).
		SetAmount(w.Amount).
		SetPaymentMethod(w.PaymentMethod).
		SetPaymentAccount(w.PaymentAccount).
		SetPaymentName(w.PaymentName).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return translatePersistenceError(err, nil, nil)
	}

	if err := tx.Commit(); err != nil {
		return translatePersistenceError(err, nil, nil)
	}

	w.ID = record.ID
	w.Status = record.Status
	w.CreatedAt = record.CreatedAt
	return nil
}

func (r *ResellerWithdrawalRepo) GetByID(ctx context.Context, id int64) (*service.ResellerWithdrawal, error) {
	record, err := r.clientFromCtx(ctx).ResellerWithdrawal.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrWithdrawalNotFound, nil)
	}
	return entToServiceWithdrawal(record), nil
}

func (r *ResellerWithdrawalRepo) List(ctx context.Context, resellerID int64, status string, limit, offset int) ([]*service.ResellerWithdrawal, int, error) {
	q := r.clientFromCtx(ctx).ResellerWithdrawal.Query().
		Where(resellerwithdrawal.ResellerIDEQ(resellerID))
	if status != "" {
		q = q.Where(resellerwithdrawal.StatusEQ(status))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, translatePersistenceError(err, nil, nil)
	}
	records, err := q.Order(dbent.Desc(resellerwithdrawal.FieldCreatedAt)).
		Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, translatePersistenceError(err, nil, nil)
	}
	result := make([]*service.ResellerWithdrawal, len(records))
	for i, rec := range records {
		result[i] = entToServiceWithdrawal(rec)
	}
	return result, total, nil
}

func (r *ResellerWithdrawalRepo) ListAll(ctx context.Context, status string, resellerID int64, limit, offset int) ([]*service.ResellerWithdrawal, int, error) {
	q := r.clientFromCtx(ctx).ResellerWithdrawal.Query()
	if status != "" {
		q = q.Where(resellerwithdrawal.StatusEQ(status))
	}
	if resellerID > 0 {
		q = q.Where(resellerwithdrawal.ResellerIDEQ(resellerID))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, translatePersistenceError(err, nil, nil)
	}
	records, err := q.Order(dbent.Desc(resellerwithdrawal.FieldCreatedAt)).
		Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, translatePersistenceError(err, nil, nil)
	}
	result := make([]*service.ResellerWithdrawal, len(records))
	for i, rec := range records {
		result[i] = entToServiceWithdrawal(rec)
	}
	return result, total, nil
}

func (r *ResellerWithdrawalRepo) SumPaidByResellerID(ctx context.Context, resellerID int64) (float64, error) {
	return r.sumByResellerIDAndStatus(ctx, resellerID, "paid")
}

func (r *ResellerWithdrawalRepo) SumPendingByResellerID(ctx context.Context, resellerID int64) (float64, error) {
	return r.sumByResellerIDAndStatus(ctx, resellerID, "pending")
}

func (r *ResellerWithdrawalRepo) sumByResellerIDAndStatus(ctx context.Context, resellerID int64, status string) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}
	err := r.clientFromCtx(ctx).ResellerWithdrawal.Query().
		Where(resellerwithdrawal.ResellerIDEQ(resellerID), resellerwithdrawal.StatusEQ(status)).
		Aggregate(dbent.As(dbent.Sum(resellerwithdrawal.FieldAmount), "sum")).
		Scan(ctx, &result)
	if err != nil {
		return 0, translatePersistenceError(err, nil, nil)
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

func (r *ResellerWithdrawalRepo) UpdateStatus(ctx context.Context, id int64, status string, adminID int64, adminNotes string) error {
	upd := r.clientFromCtx(ctx).ResellerWithdrawal.UpdateOneID(id).
		SetStatus(status).
		SetAdminNotes(adminNotes).
		SetAdminID(adminID)
	now := time.Now()
	if status == "paid" {
		upd = upd.SetPaidAt(now)
	} else if status == "rejected" {
		upd = upd.SetRejectedAt(now)
	}
	_, err := upd.Save(ctx)
	return translatePersistenceError(err, service.ErrWithdrawalNotFound, nil)
}

func (r *ResellerWithdrawalRepo) Delete(ctx context.Context, id int64) error {
	return translatePersistenceError(r.clientFromCtx(ctx).ResellerWithdrawal.DeleteOneID(id).Exec(ctx), service.ErrWithdrawalNotFound, nil)
}

func entToServiceWithdrawal(record *dbent.ResellerWithdrawal) *service.ResellerWithdrawal {
	return &service.ResellerWithdrawal{
		ID:             record.ID,
		ResellerID:     record.ResellerID,
		Amount:         record.Amount,
		Status:         record.Status,
		PaymentMethod:  record.PaymentMethod,
		PaymentAccount: record.PaymentAccount,
		PaymentName:    record.PaymentName,
		AdminNotes:     record.AdminNotes,
		AdminID:        record.AdminID,
		CreatedAt:      record.CreatedAt,
		PaidAt:         record.PaidAt,
		RejectedAt:     record.RejectedAt,
	}
}
