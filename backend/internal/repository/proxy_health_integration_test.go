//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/suite"
)

type ProxyHealthSuite struct {
	suite.Suite
	ctx  context.Context
	tx   *dbent.Tx
	repo *proxyRepository
}

func (s *ProxyHealthSuite) SetupTest() {
	s.ctx = context.Background()
	s.tx = testEntTx(s.T())
	s.repo = newProxyRepositoryWithSQL(s.tx.Client(), s.tx)
}

func TestProxyHealthSuite(t *testing.T) { suite.Run(t, new(ProxyHealthSuite)) }

func (s *ProxyHealthSuite) mkProxy(name, status, failureMode string, failureBackup *int64) int64 {
	p := &service.Proxy{Name: name, Protocol: "socks5", Host: "127.0.0.1", Port: 1080,
		Status: status, FallbackMode: service.FallbackModeNone, ExpiryWarnDays: 7,
		FailureFallbackMode: failureMode, FailureBackupProxyID: failureBackup}
	s.Require().NoError(s.repo.Create(s.ctx, p))
	return p.ID
}

func (s *ProxyHealthSuite) setDegraded(proxyID int64) {
	_, err := s.tx.ExecContext(s.ctx, `UPDATE proxies SET health_status='degraded', health_last_error='boom' WHERE id=$1`, proxyID)
	s.Require().NoError(err)
}

func (s *ProxyHealthSuite) mkAccount(proxyID, originID *int64, tempReason string) int64 {
	var until *time.Time
	var reason *string
	if tempReason != "" {
		t := time.Now().Add(10 * time.Minute)
		until, reason = &t, &tempReason
	}
	var id int64
	err := scanSingleRow(s.ctx, s.tx, `
		INSERT INTO accounts (name, platform, type, credentials, extra, status, proxy_id, proxy_fallback_origin_id,
			temp_unschedulable_until, temp_unschedulable_reason, created_at, updated_at)
		VALUES ($1,'openai','oauth','{}','{}','active',$2,$3,$4,$5,NOW(),NOW()) RETURNING id`,
		[]any{"acc-" + time.Now().Format("150405.000000000"), proxyID, originID, until, reason}, &id)
	s.Require().NoError(err)
	return id
}

type healthAccountRow struct {
	proxyID         sql.NullInt64
	originID        sql.NullInt64
	failureOriginID sql.NullInt64
	tempUntil       sql.NullTime
	tempReason      sql.NullString
}

func (s *ProxyHealthSuite) account(id int64) healthAccountRow {
	var r healthAccountRow
	s.Require().NoError(scanSingleRow(s.ctx, s.tx,
		`SELECT proxy_id, proxy_fallback_origin_id, proxy_failure_origin_id, temp_unschedulable_until, temp_unschedulable_reason FROM accounts WHERE id=$1`,
		[]any{id}, &r.proxyID, &r.originID, &r.failureOriginID, &r.tempUntil, &r.tempReason))
	return r
}

func (s *ProxyHealthSuite) outboxCount() int {
	var n int
	s.Require().NoError(scanSingleRow(s.ctx, s.tx, `SELECT COUNT(*) FROM scheduler_outbox WHERE event_type=$1`,
		[]any{service.SchedulerOutboxEventAccountBulkChanged}, &n))
	return n
}

func (s *ProxyHealthSuite) TestRecordProxyHealthResultTracksStreaks() {
	pid := s.mkProxy("streak", service.StatusActive, service.FallbackModeDirect, nil)

	st, err := s.repo.RecordProxyHealthResult(s.ctx, pid, false)
	s.Require().NoError(err)
	s.Require().Equal(service.ProxyHealthStreak{FailStreak: 1, OkStreak: 0, HealthStatus: service.ProxyHealthHealthy, Found: true}, st)
	st, err = s.repo.RecordProxyHealthResult(s.ctx, pid, false)
	s.Require().NoError(err)
	s.Require().Equal(2, st.FailStreak)
	st, err = s.repo.RecordProxyHealthResult(s.ctx, pid, true)
	s.Require().NoError(err)
	s.Require().Equal(0, st.FailStreak)
	s.Require().Equal(1, st.OkStreak)

	st, err = s.repo.RecordProxyHealthResult(s.ctx, 99999999, true)
	s.Require().NoError(err)
	s.Require().False(st.Found)
}

func (s *ProxyHealthSuite) TestListProxiesForHealthCheck() {
	withFallback := s.mkProxy("hc-direct", service.StatusActive, service.FallbackModeDirect, nil)
	noFallback := s.mkProxy("hc-none", service.StatusActive, service.FallbackModeNone, nil)
	inactive := s.mkProxy("hc-inactive", "inactive", service.FallbackModeDirect, nil)
	degradedNoFallback := s.mkProxy("hc-degraded-none", service.StatusActive, service.FallbackModeNone, nil)
	s.setDegraded(degradedNoFallback)

	list, err := s.repo.ListProxiesForHealthCheck(s.ctx)
	s.Require().NoError(err)
	ids := map[int64]bool{}
	for _, p := range list {
		ids[p.ID] = true
	}
	s.Require().True(ids[withFallback])
	s.Require().True(ids[degradedNoFallback])
	s.Require().False(ids[noFallback])
	s.Require().False(ids[inactive])
}

func (s *ProxyHealthSuite) TestMarkProxyDegradedFallsBackToDirect() {
	pid := s.mkProxy("deg-direct", service.StatusActive, service.FallbackModeDirect, nil)
	other := s.mkProxy("deg-other", service.StatusActive, service.FallbackModeNone, nil)
	transportHeld := s.mkAccount(&pid, nil, service.ProxyTransportTempUnschedReasonPrefix+" (proxy/network): dial tcp: refused")
	manualHeld := s.mkAccount(&pid, nil, "manual hold")
	untouched := s.mkAccount(&other, nil, "")
	before := s.outboxCount()

	ids, transitioned, err := s.repo.MarkProxyDegraded(s.ctx, pid, nil, true, "probe failed")
	s.Require().NoError(err)
	s.Require().True(transitioned)
	s.Require().ElementsMatch([]int64{transportHeld, manualHeld}, ids)
	s.Require().Greater(s.outboxCount(), before)

	a := s.account(transportHeld)
	s.Require().False(a.proxyID.Valid)
	s.Require().Equal(pid, a.originID.Int64)
	s.Require().Equal(pid, a.failureOriginID.Int64)
	s.Require().False(a.tempUntil.Valid, "proxy transport hold must be cleared")
	s.Require().False(a.tempReason.Valid)

	b := s.account(manualHeld)
	s.Require().False(b.proxyID.Valid)
	s.Require().True(b.tempUntil.Valid, "unrelated temp hold must be kept")
	s.Require().Equal("manual hold", b.tempReason.String)

	s.Require().Equal(other, s.account(untouched).proxyID.Int64)

	got, err := s.repo.GetByID(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().Equal(service.ProxyHealthDegraded, got.HealthStatus)
	s.Require().Equal("probe failed", got.HealthLastError)
	s.Require().NotNil(got.HealthChangedAt)
	changedAt := *got.HealthChangedAt

	// 已故障时重复调用：不算状态迁移、不改变迁移时间，只改投故障期间新绑定的账号。
	late := s.mkAccount(&pid, nil, "")
	ids, transitioned, err = s.repo.MarkProxyDegraded(s.ctx, pid, nil, true, "still failing")
	s.Require().NoError(err)
	s.Require().False(transitioned)
	s.Require().Equal([]int64{late}, ids)
	got, err = s.repo.GetByID(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().True(changedAt.Equal(*got.HealthChangedAt))
	s.Require().Equal("still failing", got.HealthLastError)
}

func (s *ProxyHealthSuite) TestChainedFailureFollowsCurrentTargetAndRestoresHome() {
	backup := s.mkProxy("chain-backup", service.StatusActive, service.FallbackModeDirect, nil)
	primary := s.mkProxy("chain-primary", service.StatusActive, service.FallbackModeProxy, &backup)
	acc := s.mkAccount(&primary, nil, "")

	_, _, err := s.repo.MarkProxyDegraded(s.ctx, primary, &backup, true, "primary down")
	s.Require().NoError(err)
	a := s.account(acc)
	s.Require().Equal(backup, a.proxyID.Int64)
	s.Require().Equal(primary, a.failureOriginID.Int64)

	// 备用代理也故障、回退直连：故障来源仍记 primary。
	_, _, err = s.repo.MarkProxyDegraded(s.ctx, backup, nil, true, "backup down")
	s.Require().NoError(err)
	a = s.account(acc)
	s.Require().False(a.proxyID.Valid)
	s.Require().Equal(primary, a.failureOriginID.Int64)

	// 备用代理恢复：它不是这个账号的故障来源，不直接改回。
	ids, transitioned, err := s.repo.MarkProxyHealthy(s.ctx, backup)
	s.Require().NoError(err)
	s.Require().True(transitioned)
	s.Require().Empty(ids)

	// primary 仍故障，下一轮重扫时目标又能解析到 backup：账号跟随最新目标回到 backup。
	ids, transitioned, err = s.repo.MarkProxyDegraded(s.ctx, primary, &backup, true, "primary still down")
	s.Require().NoError(err)
	s.Require().False(transitioned)
	s.Require().Equal([]int64{acc}, ids)
	s.Require().Equal(backup, s.account(acc).proxyID.Int64)

	// primary 恢复：改回 primary 并结束跟踪。
	ids, transitioned, err = s.repo.MarkProxyHealthy(s.ctx, primary)
	s.Require().NoError(err)
	s.Require().True(transitioned)
	s.Require().Equal([]int64{acc}, ids)
	a = s.account(acc)
	s.Require().Equal(primary, a.proxyID.Int64)
	s.Require().False(a.failureOriginID.Valid)
	s.Require().False(a.originID.Valid)
}

func (s *ProxyHealthSuite) TestFailureOnExpiryBackupRestoresToBackupAndKeepsExpiryOrigin() {
	backup := s.mkProxy("exp-backup", service.StatusActive, service.FallbackModeDirect, nil)
	expired := s.mkProxy("exp-old", service.StatusExpired, service.FallbackModeNone, nil)
	acc := s.mkAccount(&backup, &expired, "")

	_, _, err := s.repo.MarkProxyDegraded(s.ctx, backup, nil, true, "backup down")
	s.Require().NoError(err)
	a := s.account(acc)
	s.Require().False(a.proxyID.Valid)
	s.Require().Equal(expired, a.originID.Int64, "expiry origin kept for manual revert")
	s.Require().Equal(backup, a.failureOriginID.Int64)

	ids, _, err := s.repo.MarkProxyHealthy(s.ctx, backup)
	s.Require().NoError(err)
	s.Require().Equal([]int64{acc}, ids)
	a = s.account(acc)
	s.Require().Equal(backup, a.proxyID.Int64)
	s.Require().Equal(expired, a.originID.Int64)
	s.Require().False(a.failureOriginID.Valid)
}

func (s *ProxyHealthSuite) TestReSweepFollowsConfigChangedToDirect() {
	backup := s.mkProxy("cfg-backup", service.StatusActive, service.FallbackModeNone, nil)
	primary := s.mkProxy("cfg-primary", service.StatusActive, service.FallbackModeProxy, &backup)
	acc := s.mkAccount(&primary, nil, "")

	_, _, err := s.repo.MarkProxyDegraded(s.ctx, primary, &backup, true, "down")
	s.Require().NoError(err)
	s.Require().Equal(backup, s.account(acc).proxyID.Int64)

	// 管理员在故障期间把故障回退改成直连：重扫后账号跟随到直连，恢复时仍改回 primary。
	_, _, err = s.repo.MarkProxyDegraded(s.ctx, primary, nil, true, "down")
	s.Require().NoError(err)
	s.Require().False(s.account(acc).proxyID.Valid)
	_, _, err = s.repo.MarkProxyHealthy(s.ctx, primary)
	s.Require().NoError(err)
	s.Require().Equal(primary, s.account(acc).proxyID.Int64)
}

func (s *ProxyHealthSuite) TestManualRebindEndsFailureTracking() {
	manual := s.mkProxy("rebind-manual", service.StatusActive, service.FallbackModeNone, nil)
	primary := s.mkProxy("rebind-primary", service.StatusActive, service.FallbackModeDirect, nil)
	bulkAcc := s.mkAccount(&primary, nil, "")
	sameAcc := s.mkAccount(&primary, nil, "")
	_, _, err := s.repo.MarkProxyDegraded(s.ctx, primary, nil, true, "down")
	s.Require().NoError(err)

	accRepo := newAccountRepositoryWithSQL(s.tx.Client(), s.tx, nil)
	_, err = accRepo.BulkUpdate(s.ctx, []int64{bulkAcc}, service.AccountBulkUpdate{ProxyID: &manual})
	s.Require().NoError(err)
	a := s.account(bulkAcc)
	s.Require().Equal(manual, a.proxyID.Int64)
	s.Require().False(a.failureOriginID.Valid)

	// 批量编辑没改代理（仍是直连 0）时不结束跟踪。
	zero := int64(0)
	_, err = accRepo.BulkUpdate(s.ctx, []int64{sameAcc}, service.AccountBulkUpdate{ProxyID: &zero})
	s.Require().NoError(err)
	s.Require().Equal(primary, s.account(sameAcc).failureOriginID.Int64)

	ids, _, err := s.repo.MarkProxyHealthy(s.ctx, primary)
	s.Require().NoError(err)
	s.Require().Equal([]int64{sameAcc}, ids)
	s.Require().Equal(manual, s.account(bulkAcc).proxyID.Int64)
}

func (s *ProxyHealthSuite) mkShadow(parentID int64, proxyID *int64) int64 {
	var id int64
	err := scanSingleRow(s.ctx, s.tx, `
		INSERT INTO accounts (name, platform, type, credentials, extra, status, proxy_id, parent_account_id, quota_dimension, created_at, updated_at)
		VALUES ($1,'openai','oauth','{}','{}','active',$2,$3,'spark',NOW(),NOW()) RETURNING id`,
		[]any{"shadow-" + time.Now().Format("150405.000000000"), proxyID, parentID}, &id)
	s.Require().NoError(err)
	return id
}

func (s *ProxyHealthSuite) TestShadowCreatedDuringFallbackFollowsParentBack() {
	backup := s.mkProxy("shadow-backup", service.StatusActive, service.FallbackModeNone, nil)
	primary := s.mkProxy("shadow-primary", service.StatusActive, service.FallbackModeProxy, &backup)
	parent := s.mkAccount(&primary, nil, "")
	_, _, err := s.repo.MarkProxyDegraded(s.ctx, primary, &backup, true, "down")
	s.Require().NoError(err)

	// 故障期间新建的影子复制了母账号当前代理（backup），但不带故障回退来源。
	shadow := s.mkShadow(parent, &backup)

	ids, _, err := s.repo.MarkProxyHealthy(s.ctx, primary)
	s.Require().NoError(err)
	s.Require().ElementsMatch([]int64{parent, shadow}, ids)
	s.Require().Equal(primary, s.account(shadow).proxyID.Int64)
	s.Require().False(s.account(shadow).failureOriginID.Valid)
}

func (s *ProxyHealthSuite) TestDeleteProxyClearsFailureOrigin() {
	primary := s.mkProxy("del-primary", service.StatusActive, service.FallbackModeDirect, nil)
	acc := s.mkAccount(&primary, nil, "")
	_, _, err := s.repo.MarkProxyDegraded(s.ctx, primary, nil, true, "down")
	s.Require().NoError(err)
	s.Require().NoError(s.repo.Delete(s.ctx, primary))
	s.Require().False(s.account(acc).failureOriginID.Valid)
}

func (s *ProxyHealthSuite) TestManualRevertEndsFailureTracking() {
	primary := s.mkProxy("revert-primary", service.StatusActive, service.FallbackModeDirect, nil)
	acc := s.mkAccount(&primary, nil, "")
	_, _, err := s.repo.MarkProxyDegraded(s.ctx, primary, nil, true, "down")
	s.Require().NoError(err)

	accRepo := newAccountRepositoryWithSQL(s.tx.Client(), s.tx, nil)
	s.Require().NoError(accRepo.RevertProxyFallback(s.ctx, acc))
	a := s.account(acc)
	s.Require().Equal(primary, a.proxyID.Int64)
	s.Require().False(a.failureOriginID.Valid)
	s.Require().False(a.originID.Valid)
}

func (s *ProxyHealthSuite) TestMarkProxyDegradedWithoutChangeOnlyMarksProxy() {
	pid := s.mkProxy("deg-nochange", service.StatusActive, service.FallbackModeProxy, nil)
	acc := s.mkAccount(&pid, nil, "")

	ids, transitioned, err := s.repo.MarkProxyDegraded(s.ctx, pid, nil, false, "chain unresolved")
	s.Require().NoError(err)
	s.Require().True(transitioned)
	s.Require().Empty(ids)
	s.Require().Equal(pid, s.account(acc).proxyID.Int64)
}

func (s *ProxyHealthSuite) TestMarkProxyHealthyOnlyActsOnDegradedProxy() {
	pid := s.mkProxy("heal-healthy", service.StatusActive, service.FallbackModeDirect, nil)
	stray := s.mkAccount(nil, &pid, "")

	ids, transitioned, err := s.repo.MarkProxyHealthy(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().False(transitioned)
	s.Require().Empty(ids)
	s.Require().False(s.account(stray).proxyID.Valid, "expiry-origin accounts are not touched by failure recovery")

	s.setDegraded(pid)
	ids, transitioned, err = s.repo.MarkProxyHealthy(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().True(transitioned)
	s.Require().Empty(ids, "only accounts moved by failure fallback are restored")
	got, err := s.repo.GetByID(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().Equal(service.ProxyHealthHealthy, got.HealthStatus)
	s.Require().Empty(got.HealthLastError)
}

func (s *ProxyHealthSuite) TestUpdatePersistsFailureFallback() {
	backup := s.mkProxy("upd-backup", service.StatusActive, service.FallbackModeNone, nil)
	pid := s.mkProxy("upd", service.StatusActive, service.FallbackModeNone, nil)
	p, err := s.repo.GetByID(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().Equal(service.FallbackModeNone, p.FailureFallbackMode)
	s.Require().Equal(service.ProxyHealthHealthy, p.HealthStatus)

	p.FailureFallbackMode = service.FallbackModeProxy
	p.FailureBackupProxyID = &backup
	s.Require().NoError(s.repo.Update(s.ctx, p))
	got, err := s.repo.GetByID(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().Equal(service.FallbackModeProxy, got.FailureFallbackMode)
	s.Require().Equal(backup, *got.FailureBackupProxyID)

	_, err = s.repo.RecordProxyHealthResult(s.ctx, pid, false)
	s.Require().NoError(err)
	got.FailureFallbackMode = service.FallbackModeDirect
	got.FailureBackupProxyID = nil
	s.Require().NoError(s.repo.Update(s.ctx, got))
	st, err := s.repo.RecordProxyHealthResult(s.ctx, pid, false)
	s.Require().NoError(err)
	s.Require().Equal(1, st.FailStreak, "admin update restarts health streaks")
	got, err = s.repo.GetByID(s.ctx, pid)
	s.Require().NoError(err)
	s.Require().Equal(service.FallbackModeDirect, got.FailureFallbackMode)
	s.Require().Nil(got.FailureBackupProxyID)
}
