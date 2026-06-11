package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- 最小有状态 fake：充值订单仓储（EPAY / AliMPay 共用同样的 CAS 语义） ---

type fakeRechargeOrderRepo struct {
	RechargeOrderRepository // 嵌入接口，未实现的方法在被调用时 panic（测试不会触及）
	order                   *RechargeOrder
	getErr                  error
	updateCalls             [][2]string // 记录每次 (fromStatus,toStatus)
}

func (f *fakeRechargeOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*RechargeOrder, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.order == nil || f.order.OrderNo != orderNo {
		return nil, nil
	}
	cp := *f.order
	return &cp, nil
}

func (f *fakeRechargeOrderRepo) UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, tradeNo *string, paidAt *time.Time) error {
	f.updateCalls = append(f.updateCalls, [2]string{fromStatus, toStatus})
	if f.order == nil || f.order.OrderNo != orderNo || f.order.Status != fromStatus {
		return ErrRechargeOrderStatusConflict // CAS 失败
	}
	f.order.Status = toStatus
	return nil
}

type fakeAliMPayOrderRepo struct {
	OrderRepository
	order       *Order
	updateCalls [][2]string
}

func (f *fakeAliMPayOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	if f.order == nil || f.order.OrderNo != orderNo {
		return nil, nil
	}
	cp := *f.order
	return &cp, nil
}

func (f *fakeAliMPayOrderRepo) UpdateStatus(ctx context.Context, orderNo, fromStatus, toStatus string, tradeNo *string, paidAt *time.Time) error {
	f.updateCalls = append(f.updateCalls, [2]string{fromStatus, toStatus})
	if f.order == nil || f.order.OrderNo != orderNo || f.order.Status != fromStatus {
		return ErrOrderStatusConflict
	}
	f.order.Status = toStatus
	return nil
}

// --- 最小 fake：AdminService，仅实现退款扣回 ---

type fakeRefundAdminService struct {
	AdminService
	refundErr      error
	refundedUserID int64
	refundedAmount float64
	refundCalls    int
}

func (f *fakeRefundAdminService) RefundUserBalance(ctx context.Context, userID int64, amount float64, notes string) error {
	f.refundCalls++
	if f.refundErr != nil {
		return f.refundErr
	}
	f.refundedUserID = userID
	f.refundedAmount = amount
	return nil
}

// ---------- EPAY RechargeService.RefundOrder ----------

func TestRechargeRefundOrder_Success(t *testing.T) {
	repo := &fakeRechargeOrderRepo{order: &RechargeOrder{
		OrderNo: "R1", UserID: 42, CreditAmount: 12.5, Status: "paid",
	}}
	admin := &fakeRefundAdminService{}
	svc := &RechargeService{orderRepo: repo, adminService: admin}

	order, err := svc.RefundOrder(context.Background(), "R1", "user requested")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != "refunded" {
		t.Fatalf("returned order status = %q, want refunded", order.Status)
	}
	if repo.order.Status != "refunded" {
		t.Fatalf("repo order status = %q, want refunded (commission base must drop)", repo.order.Status)
	}
	if admin.refundCalls != 1 || admin.refundedUserID != 42 || admin.refundedAmount != 12.5 {
		t.Fatalf("clawback not invoked correctly: calls=%d user=%d amount=%.2f", admin.refundCalls, admin.refundedUserID, admin.refundedAmount)
	}
	if len(repo.updateCalls) != 1 || repo.updateCalls[0] != [2]string{"paid", "refunded"} {
		t.Fatalf("unexpected status transitions: %v", repo.updateCalls)
	}
}

func TestRechargeRefundOrder_NotPaid(t *testing.T) {
	repo := &fakeRechargeOrderRepo{order: &RechargeOrder{OrderNo: "R1", Status: "pending"}}
	admin := &fakeRefundAdminService{}
	svc := &RechargeService{orderRepo: repo, adminService: admin}

	_, err := svc.RefundOrder(context.Background(), "R1", "")
	if !errors.Is(err, ErrRechargeOrderNotRefundable) {
		t.Fatalf("err = %v, want ErrRechargeOrderNotRefundable", err)
	}
	if admin.refundCalls != 0 {
		t.Fatalf("clawback must not run for non-paid order")
	}
	if repo.order.Status != "pending" {
		t.Fatalf("status changed unexpectedly: %q", repo.order.Status)
	}
}

func TestRechargeRefundOrder_NotFound(t *testing.T) {
	repo := &fakeRechargeOrderRepo{order: nil}
	svc := &RechargeService{orderRepo: repo, adminService: &fakeRefundAdminService{}}

	_, err := svc.RefundOrder(context.Background(), "missing", "")
	if !errors.Is(err, ErrRechargeOrderNotFound) {
		t.Fatalf("err = %v, want ErrRechargeOrderNotFound", err)
	}
}

func TestRechargeRefundOrder_ClawbackFailureCompensates(t *testing.T) {
	repo := &fakeRechargeOrderRepo{order: &RechargeOrder{
		OrderNo: "R1", UserID: 7, CreditAmount: 30, Status: "paid",
	}}
	admin := &fakeRefundAdminService{refundErr: errors.New("db down")}
	svc := &RechargeService{orderRepo: repo, adminService: admin}

	_, err := svc.RefundOrder(context.Background(), "R1", "")
	if err == nil {
		t.Fatalf("expected error when clawback fails")
	}
	// 补偿：订单状态应回滚为 paid，避免"已退款但余额未扣回"。
	if repo.order.Status != "paid" {
		t.Fatalf("status = %q, want rolled back to paid", repo.order.Status)
	}
	wantCalls := [][2]string{{"paid", "refunded"}, {"refunded", "paid"}}
	if len(repo.updateCalls) != 2 || repo.updateCalls[0] != wantCalls[0] || repo.updateCalls[1] != wantCalls[1] {
		t.Fatalf("compensation transitions = %v, want %v", repo.updateCalls, wantCalls)
	}
}

func TestRechargeRefundOrder_CASConflict(t *testing.T) {
	// GetByOrderNo 读到 paid，但 UpdateStatus 前并发改成了 refunded → CAS 冲突。
	repo := &fakeRechargeOrderRepo{order: &RechargeOrder{OrderNo: "R1", Status: "paid"}}
	admin := &fakeRefundAdminService{}
	svc := &RechargeService{orderRepo: repo, adminService: admin}
	// 模拟并发：在 Get 之后、Update 之前把状态改掉。
	repo.order.Status = "refunded"

	_, err := svc.RefundOrder(context.Background(), "R1", "")
	if !errors.Is(err, ErrRechargeOrderNotRefundable) {
		t.Fatalf("err = %v, want ErrRechargeOrderNotRefundable on CAS conflict", err)
	}
	if admin.refundCalls != 0 {
		t.Fatalf("clawback must not run when CAS lost")
	}
}

// ---------- AliMPay OrderService.RefundOrder ----------

func TestAliMPayRefundOrder_Success(t *testing.T) {
	repo := &fakeAliMPayOrderRepo{order: &Order{OrderNo: "A1", UserID: 9, CreditAmount: 88, Status: "paid"}}
	admin := &fakeRefundAdminService{}
	svc := &OrderService{orderRepo: repo, adminService: admin}

	order, err := svc.RefundOrder(context.Background(), "A1", "dup pay")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != "refunded" || repo.order.Status != "refunded" {
		t.Fatalf("status not refunded: ret=%q repo=%q", order.Status, repo.order.Status)
	}
	if admin.refundCalls != 1 || admin.refundedAmount != 88 {
		t.Fatalf("clawback wrong: calls=%d amount=%.2f", admin.refundCalls, admin.refundedAmount)
	}
}

func TestAliMPayRefundOrder_ClawbackFailureCompensates(t *testing.T) {
	repo := &fakeAliMPayOrderRepo{order: &Order{OrderNo: "A1", UserID: 9, CreditAmount: 88, Status: "paid"}}
	admin := &fakeRefundAdminService{refundErr: errors.New("boom")}
	svc := &OrderService{orderRepo: repo, adminService: admin}

	if _, err := svc.RefundOrder(context.Background(), "A1", ""); err == nil {
		t.Fatalf("expected error")
	}
	if repo.order.Status != "paid" {
		t.Fatalf("status = %q, want rolled back to paid", repo.order.Status)
	}
}

func TestAliMPayRefundOrder_NotPaid(t *testing.T) {
	repo := &fakeAliMPayOrderRepo{order: &Order{OrderNo: "A1", Status: "expired"}}
	svc := &OrderService{orderRepo: repo, adminService: &fakeRefundAdminService{}}

	if _, err := svc.RefundOrder(context.Background(), "A1", ""); !errors.Is(err, ErrOrderNotRefundable) {
		t.Fatalf("err = %v, want ErrOrderNotRefundable", err)
	}
}

func TestAliMPayRefundOrder_NotFound(t *testing.T) {
	repo := &fakeAliMPayOrderRepo{order: nil}
	svc := &OrderService{orderRepo: repo, adminService: &fakeRefundAdminService{}}

	if _, err := svc.RefundOrder(context.Background(), "missing", ""); !errors.Is(err, ErrAliMPayOrderNotFound) {
		t.Fatalf("err = %v, want ErrAliMPayOrderNotFound", err)
	}
}

// ---------- adminServiceImpl.RefundUserBalance ----------

type refundUserRepo struct {
	UserRepository
	balance     float64
	deductCalls []float64
}

func (r *refundUserRepo) DeductBalance(ctx context.Context, id int64, amount float64) error {
	r.deductCalls = append(r.deductCalls, amount)
	r.balance -= amount // 允许透支：不做余额充足校验
	return nil
}

func (r *refundUserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	return &User{ID: id, Balance: r.balance, Status: StatusActive}, nil
}

type refundRedeemRepo struct {
	RedeemCodeRepository
	created []*RedeemCode
}

func (r *refundRedeemRepo) Create(ctx context.Context, code *RedeemCode) error {
	clone := *code
	r.created = append(r.created, &clone)
	return nil
}

func TestRefundUserBalance_AllowsOverdraftAndAudits(t *testing.T) {
	// 用户余额 5，但要扣回 12.5（已消费部分）→ 余额应透支为 -7.5。
	userRepo := &refundUserRepo{balance: 5}
	redeemRepo := &refundRedeemRepo{}
	svc := &adminServiceImpl{userRepo: userRepo, redeemCodeRepo: redeemRepo}

	if err := svc.RefundUserBalance(context.Background(), 42, 12.5, "Refund recharge order R1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(userRepo.deductCalls) != 1 || userRepo.deductCalls[0] != 12.5 {
		t.Fatalf("DeductBalance calls = %v, want [12.5]", userRepo.deductCalls)
	}
	if userRepo.balance != -7.5 {
		t.Fatalf("balance = %.2f, want -7.5 (overdraft allowed)", userRepo.balance)
	}
	if len(redeemRepo.created) != 1 {
		t.Fatalf("expected one audit record, got %d", len(redeemRepo.created))
	}
	rec := redeemRepo.created[0]
	if rec.Value != -12.5 {
		t.Fatalf("audit value = %.2f, want -12.5 (negative, excluded from inflow)", rec.Value)
	}
	if rec.Type != AdjustmentTypeAdminBalance || rec.Status != StatusUsed {
		t.Fatalf("audit record type/status = %q/%q", rec.Type, rec.Status)
	}
}

func TestRefundUserBalance_RejectsNonPositive(t *testing.T) {
	svc := &adminServiceImpl{userRepo: &refundUserRepo{}, redeemCodeRepo: &refundRedeemRepo{}}
	if err := svc.RefundUserBalance(context.Background(), 1, 0, ""); err == nil {
		t.Fatalf("expected error for non-positive refund amount")
	}
}
