package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	failoverSoftReconcileInterval = 30 * time.Second
	failoverRecoveryProbeInterval = 30 * time.Second
	failoverEventRetention        = 90 * 24 * time.Hour
	failoverEventCleanupInterval  = 24 * time.Hour
	failoverSoftReconcileLockKey  = "leader:failover_soft_reconcile"
	failoverRecoveryProbeLockKey  = "leader:failover_recovery_prober"
	failoverEventCleanupLockKey   = "leader:failover_event_cleanup"
	failoverProbeTimeout          = 60 * time.Second
)

type failoverActiveTransitionWriter interface {
	TransitionFailoverActive(ctx context.Context, groupID int64, newMemberID int64, expectedVersion int64, event *FailoverGroupEvent) (bool, error)
}

type failoverPinStateWriter interface {
	UpdateFailoverPinState(ctx context.Context, groupID int64, memberID *int64, expiresAt *time.Time, event *FailoverGroupEvent) error
}

// FailoverGroupService 负责维护"智能路由"虚拟分组的激活成员。
//
// 三个独立的 reconciler 共同驱动状态迁移：
//  1. 软状态降级（30 秒 tick）：当前 active 已无可调度账号时，向后寻找第一个可用成员
//  2. 恢复探测（30 秒 tick，仅 active != members[0] 时运行）：轻量探测更高优先级成员
//  3. 健康检查 reconcile 钩子：5 分钟一次，由 GroupHealthCheckService 触发
//
// 所有状态写入均使用乐观锁（failover_active_version）并追加一条 failover_group_events。
type FailoverGroupService struct {
	groupRepo      GroupRepository
	accountRepo    AccountRepository
	eventRepo      FailoverEventRepository
	accountTestSvc *AccountTestService
	locker         *LeaderLocker

	softStop    chan struct{}
	probeStop   chan struct{}
	cleanupStop chan struct{}
	startOnce   sync.Once
	stopOnce    sync.Once
}

func NewFailoverGroupService(
	groupRepo GroupRepository,
	accountRepo AccountRepository,
	eventRepo FailoverEventRepository,
	accountTestSvc *AccountTestService,
	locker *LeaderLocker,
) *FailoverGroupService {
	return &FailoverGroupService{
		groupRepo:      groupRepo,
		accountRepo:    accountRepo,
		eventRepo:      eventRepo,
		accountTestSvc: accountTestSvc,
		locker:         locker,
		softStop:       make(chan struct{}),
		probeStop:      make(chan struct{}),
		cleanupStop:    make(chan struct{}),
	}
}

// Start 启动两个 ticker，内部 goroutine 通过 LeaderLocker 做集群唯一。
func (s *FailoverGroupService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		go s.runSoftReconcileLoop()
		go s.runRecoveryProbeLoop()
		go s.runEventCleanupLoop()
		logger.LegacyPrintf("service.failover_group", "[FailoverGroup] started (soft_tick=30s, probe_tick=30s, cleanup_tick=24h)")
	})
}

// Stop 停止 ticker。
func (s *FailoverGroupService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.softStop)
		close(s.probeStop)
		close(s.cleanupStop)
	})
}

func (s *FailoverGroupService) runSoftReconcileLoop() {
	ticker := time.NewTicker(failoverSoftReconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.softStop:
			return
		case <-ticker.C:
			s.runSoftReconcileOnce()
		}
	}
}

func (s *FailoverGroupService) runEventCleanupLoop() {
	// 启动后延迟 5 分钟再做第一次，避免与启动期其他 IO 竞争
	initialDelay := time.NewTimer(5 * time.Minute)
	select {
	case <-s.cleanupStop:
		initialDelay.Stop()
		return
	case <-initialDelay.C:
	}
	s.runEventCleanupOnce()
	ticker := time.NewTicker(failoverEventCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.cleanupStop:
			return
		case <-ticker.C:
			s.runEventCleanupOnce()
		}
	}
}

func (s *FailoverGroupService) runEventCleanupOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	release, ok := s.locker.TryAcquire(ctx, failoverEventCleanupLockKey, 5*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}
	cutoff := time.Now().Add(-failoverEventRetention)
	n, err := s.eventRepo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		logger.LegacyPrintf("service.failover_group", "[FailoverGroup] event cleanup failed: %v", err)
		return
	}
	if n > 0 {
		logger.LegacyPrintf("service.failover_group", "[FailoverGroup] event cleanup deleted=%d cutoff=%s", n, cutoff.UTC().Format(time.RFC3339))
	}
}

func (s *FailoverGroupService) runRecoveryProbeLoop() {
	ticker := time.NewTicker(failoverRecoveryProbeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.probeStop:
			return
		case <-ticker.C:
			s.runRecoveryProbeOnce()
		}
	}
}

func (s *FailoverGroupService) runSoftReconcileOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	release, ok := s.locker.TryAcquire(ctx, failoverSoftReconcileLockKey, 2*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}
	groups, err := s.groupRepo.ListFailoverGroups(ctx)
	if err != nil {
		logger.LegacyPrintf("service.failover_group", "[FailoverGroup] list for soft reconcile failed: %v", err)
		return
	}
	for _, g := range groups {
		if err := s.softReconcileOne(ctx, g); err != nil {
			logger.LegacyPrintf("service.failover_group", "[FailoverGroup] soft reconcile group=%d err=%v", g.ID, err)
		}
	}
}

func (s *FailoverGroupService) runRecoveryProbeOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	release, ok := s.locker.TryAcquire(ctx, failoverRecoveryProbeLockKey, 3*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	groups, err := s.groupRepo.ListFailoverGroups(ctx)
	if err != nil {
		logger.LegacyPrintf("service.failover_group", "[FailoverGroup] list for probe failed: %v", err)
		return
	}
	now := time.Now()
	for _, g := range groups {
		if g == nil || len(g.FailoverMemberIDs) <= 1 {
			continue
		}
		if g.IsFailoverPinActive(now) {
			continue
		}
		activePtr := g.EffectiveFailoverActiveMemberID(now)
		if activePtr == nil || *activePtr == g.FailoverMemberIDs[0] {
			continue
		}
		currentIdx := indexOfInt64(g.FailoverMemberIDs, *activePtr)
		if currentIdx <= 0 {
			continue
		}
		for i := 0; i < currentIdx; i++ {
			candidateID := g.FailoverMemberIDs[i]
			if s.probeMemberOnce(ctx, g, candidateID) {
				if err := s.switchActive(ctx, g, candidateID, FailoverEventReasonProbeRecovered, nil, nil); err != nil {
					logger.LegacyPrintf("service.failover_group", "[FailoverGroup] probe_recovered switch failed: group=%d to=%d err=%v", g.ID, candidateID, err)
				}
				break
			}
		}
	}
}

// probeMemberOnce 对成员分组的任意可调度账号触发一次轻量测试，返回是否成功。
// 直接复用 AccountTestService.RunTestBackground，单账号 60s 超时；一个成功即判定成员恢复。
func (s *FailoverGroupService) probeMemberOnce(ctx context.Context, g *Group, memberID int64) bool {
	if s.accountTestSvc == nil || s.accountRepo == nil {
		return false
	}
	count, err := s.groupRepo.CountSchedulableAccountsByGroup(ctx, memberID)
	if err != nil || count == 0 {
		return false
	}
	member, err := s.groupRepo.GetByIDLite(ctx, memberID)
	if err != nil || member == nil {
		return false
	}
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, memberID)
	if err != nil || len(accounts) == 0 {
		return false
	}
	modelID := getDefaultTestModel(member.Platform)
	for _, acc := range accounts {
		testCtx, cancel := context.WithTimeout(ctx, failoverProbeTimeout)
		result, err := s.accountTestSvc.RunTestBackground(testCtx, acc.ID, modelID)
		cancel()
		if err != nil || result == nil {
			continue
		}
		if result.Status == "success" {
			return true
		}
	}
	return false
}

// ReconcileAll 遍历所有虚拟分组执行 reconcile；triggeredBy=0 表示系统触发。
func (s *FailoverGroupService) ReconcileAll(ctx context.Context, triggeredBy int64) error {
	return s.reconcileAllWithReason(ctx, FailoverEventReasonHealthReconcile, triggeredBy)
}

func (s *FailoverGroupService) reconcileAllWithReason(ctx context.Context, reason string, triggeredBy int64) error {
	groups, err := s.groupRepo.ListFailoverGroups(ctx)
	if err != nil {
		return err
	}
	for _, g := range groups {
		if err := s.reconcileOne(ctx, g, reason, triggeredBy); err != nil {
			logger.LegacyPrintf("service.failover_group", "[FailoverGroup] reconcile group=%d err=%v", g.ID, err)
		}
	}
	return nil
}

// ForceReconcile 由 admin 配置变更或手动解锁后立即调用。
func (s *FailoverGroupService) ForceReconcile(ctx context.Context, virtualGroupID int64, triggeredBy int64) error {
	g, err := s.groupRepo.GetByIDLite(ctx, virtualGroupID)
	if err != nil {
		return err
	}
	if g == nil || !g.IsFailoverGroup {
		return nil
	}
	return s.reconcileOne(ctx, g, FailoverEventReasonConfigChange, triggeredBy)
}

func (s *FailoverGroupService) softReconcileOne(ctx context.Context, g *Group) error {
	if g == nil || !g.IsFailoverGroup || len(g.FailoverMemberIDs) == 0 {
		return nil
	}

	now := time.Now()
	reloaded, err := s.handleExpiredPin(ctx, g, now)
	if err != nil {
		return err
	}
	if reloaded != nil {
		g = reloaded
	}
	if g.IsFailoverPinActive(now) {
		return nil
	}

	activePtr := g.EffectiveFailoverActiveMemberID(now)
	if activePtr == nil {
		return nil
	}
	currentActive := *activePtr
	currentIdx := indexOfInt64(g.FailoverMemberIDs, currentActive)
	if currentIdx < 0 {
		return s.reconcileOne(ctx, g, FailoverEventReasonSoftDemote, 0)
	}

	currentCount, err := s.groupRepo.CountSchedulableAccountsByGroup(ctx, currentActive)
	if err != nil {
		return err
	}
	if currentCount > 0 {
		return nil
	}

	for i := currentIdx + 1; i < len(g.FailoverMemberIDs); i++ {
		memberID := g.FailoverMemberIDs[i]
		count, err := s.groupRepo.CountSchedulableAccountsByGroup(ctx, memberID)
		if err != nil {
			return err
		}
		if count == 0 {
			continue
		}
		reason := FailoverEventReasonSoftDemote
		if g.FailoverActiveMemberID == nil {
			reason = FailoverEventReasonBootstrap
		}
		return s.switchActive(ctx, g, memberID, reason, nil, nil)
	}

	return nil
}

func (s *FailoverGroupService) reconcileOne(ctx context.Context, g *Group, reason string, triggeredBy int64) error {
	if g == nil || !g.IsFailoverGroup || len(g.FailoverMemberIDs) == 0 {
		return nil
	}

	now := time.Now()

	reloaded, err := s.handleExpiredPin(ctx, g, now)
	if err != nil {
		return err
	}
	if reloaded != nil {
		g = reloaded
	}

	// 2. pin 活跃期间，强制 active = pin 目标
	if g.IsFailoverPinActive(now) {
		pin := *g.FailoverPinMemberID
		if g.FailoverActiveMemberID == nil || *g.FailoverActiveMemberID != pin {
			return s.syncActiveNoEvent(ctx, g, pin)
		}
		return nil
	}

	// 3. 按顺序寻找第一个可用成员
	chosen := int64(0)
	for _, memberID := range g.FailoverMemberIDs {
		count, err := s.groupRepo.CountSchedulableAccountsByGroup(ctx, memberID)
		if err != nil {
			return err
		}
		if count == 0 {
			continue
		}
		member, err := s.groupRepo.GetByIDLite(ctx, memberID)
		if err != nil || member == nil {
			continue
		}
		if member.HealthStatus == HealthStatusUnavailable {
			continue
		}
		chosen = memberID
		break
	}

	if chosen == 0 {
		// 没有任何成员可用，保持当前 active 不动，仅记录一次事件（避免刷屏：仅当最近一次不同才记录）
		if g.FailoverActiveMemberID != nil {
			return nil
		}
		return s.writeEvent(ctx, g, nil, nil, FailoverEventReasonNoMembersAvailable, nil, nil)
	}

	current := int64(-1)
	if g.FailoverActiveMemberID != nil {
		current = *g.FailoverActiveMemberID
	}
	if chosen == current {
		return nil
	}
	if g.FailoverActiveMemberID == nil {
		reason = FailoverEventReasonBootstrap
	}
	return s.switchActive(ctx, g, chosen, reason, nil, triggerByPtr(triggeredBy))
}

func (s *FailoverGroupService) handleExpiredPin(ctx context.Context, g *Group, now time.Time) (*Group, error) {
	if g == nil || g.FailoverPinMemberID == nil || g.FailoverPinExpiresAt == nil || g.FailoverPinExpiresAt.After(now) {
		return nil, nil
	}
	event := &FailoverGroupEvent{
		VirtualGroupID: g.ID,
		FromMemberID:   g.FailoverActiveMemberID,
		ToMemberID:     g.FailoverActiveMemberID,
		Reason:         FailoverEventReasonManualUnpinExpired,
	}
	if err := s.updatePinState(ctx, g.ID, nil, nil, event); err != nil {
		return nil, err
	}
	reloaded, err := s.groupRepo.GetByIDLite(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	return reloaded, nil
}

// switchActive 在 CAS 成功时写入事件；CAS 抢输视为另一个节点已处理，静默返回。
// CAS 成功后会同步刷新 in-memory 的 active member id 与 version，避免同一个 g 快照被二次 CAS。
func (s *FailoverGroupService) switchActive(ctx context.Context, g *Group, toMemberID int64, reason string, note *string, triggeredBy *int64) error {
	prev := g.FailoverActiveMemberID
	event := &FailoverGroupEvent{
		VirtualGroupID: g.ID,
		FromMemberID:   prev,
		ToMemberID:     &toMemberID,
		Reason:         reason,
		Note:           note,
		TriggeredBy:    triggeredBy,
	}
	ok, err := s.transitionActive(ctx, g.ID, toMemberID, g.FailoverActiveVersion, event)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	to := toMemberID
	g.FailoverActiveMemberID = &to
	g.FailoverActiveVersion++
	return nil
}

func (s *FailoverGroupService) writeEvent(ctx context.Context, g *Group, from *int64, to *int64, reason string, note *string, triggeredBy *int64) error {
	ev := &FailoverGroupEvent{
		VirtualGroupID: g.ID,
		FromMemberID:   from,
		ToMemberID:     to,
		Reason:         reason,
		Note:           note,
		TriggeredBy:    triggeredBy,
	}
	return s.eventRepo.Create(ctx, ev)
}

func (s *FailoverGroupService) transitionActive(ctx context.Context, groupID int64, toMemberID int64, expectedVersion int64, event *FailoverGroupEvent) (bool, error) {
	if writer, ok := s.groupRepo.(failoverActiveTransitionWriter); ok {
		return writer.TransitionFailoverActive(ctx, groupID, toMemberID, expectedVersion, event)
	}
	ok, err := s.groupRepo.UpdateFailoverActive(ctx, groupID, toMemberID, expectedVersion)
	if err != nil || !ok {
		return ok, err
	}
	if event != nil {
		g := &Group{ID: groupID}
		return true, s.writeEvent(ctx, g, event.FromMemberID, event.ToMemberID, event.Reason, event.Note, event.TriggeredBy)
	}
	return true, nil
}

func (s *FailoverGroupService) syncActiveNoEvent(ctx context.Context, g *Group, toMemberID int64) error {
	ok, err := s.groupRepo.UpdateFailoverActive(ctx, g.ID, toMemberID, g.FailoverActiveVersion)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	to := toMemberID
	g.FailoverActiveMemberID = &to
	g.FailoverActiveVersion++
	return nil
}

func (s *FailoverGroupService) updatePinState(ctx context.Context, groupID int64, memberID *int64, expiresAt *time.Time, event *FailoverGroupEvent) error {
	if writer, ok := s.groupRepo.(failoverPinStateWriter); ok {
		return writer.UpdateFailoverPinState(ctx, groupID, memberID, expiresAt, event)
	}
	var err error
	if memberID != nil {
		err = s.groupRepo.SetFailoverPin(ctx, groupID, *memberID, expiresAt)
	} else {
		err = s.groupRepo.ClearFailoverPin(ctx, groupID)
	}
	if err != nil {
		return err
	}
	if event != nil {
		g := &Group{ID: groupID}
		return s.writeEvent(ctx, g, event.FromMemberID, event.ToMemberID, event.Reason, event.Note, event.TriggeredBy)
	}
	return nil
}

// SetManualPin 手动锁定某个成员；ttl<=0 表示永久。
func (s *FailoverGroupService) SetManualPin(ctx context.Context, virtualGroupID int64, memberID int64, ttl time.Duration, adminUserID int64) error {
	g, err := s.groupRepo.GetByIDLite(ctx, virtualGroupID)
	if err != nil {
		return err
	}
	if g == nil || !g.IsFailoverGroup {
		return fmt.Errorf("group is not a failover group")
	}
	if !failoverContainsInt64(g.FailoverMemberIDs, memberID) {
		return fmt.Errorf("member %d is not part of failover group %d", memberID, virtualGroupID)
	}
	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}
	note := fmt.Sprintf("ttl=%s", ttl.String())
	event := &FailoverGroupEvent{
		VirtualGroupID: g.ID,
		FromMemberID:   g.FailoverActiveMemberID,
		ToMemberID:     &memberID,
		Reason:         FailoverEventReasonManualPin,
		Note:           &note,
		TriggeredBy:    triggerByPtr(adminUserID),
	}
	if err := s.updatePinState(ctx, virtualGroupID, &memberID, expiresAt, event); err != nil {
		return err
	}
	return s.ForceReconcile(ctx, virtualGroupID, adminUserID)
}

// ClearManualPin 清除手动锁定。
func (s *FailoverGroupService) ClearManualPin(ctx context.Context, virtualGroupID int64, adminUserID int64) error {
	g, err := s.groupRepo.GetByIDLite(ctx, virtualGroupID)
	if err != nil {
		return err
	}
	if g == nil || !g.IsFailoverGroup {
		return fmt.Errorf("group is not a failover group")
	}
	event := &FailoverGroupEvent{
		VirtualGroupID: g.ID,
		FromMemberID:   g.FailoverActiveMemberID,
		ToMemberID:     g.FailoverActiveMemberID,
		Reason:         FailoverEventReasonManualUnpin,
		TriggeredBy:    triggerByPtr(adminUserID),
	}
	if err := s.updatePinState(ctx, virtualGroupID, nil, nil, event); err != nil {
		return err
	}
	return s.ForceReconcile(ctx, virtualGroupID, adminUserID)
}

// TriggerMemberProbe 同步对指定成员触发一次轻量探测；返回成功/失败。
func (s *FailoverGroupService) TriggerMemberProbe(ctx context.Context, virtualGroupID int64, memberID int64) (bool, error) {
	g, err := s.groupRepo.GetByIDLite(ctx, virtualGroupID)
	if err != nil {
		return false, err
	}
	if g == nil || !g.IsFailoverGroup {
		return false, fmt.Errorf("group is not a failover group")
	}
	if !failoverContainsInt64(g.FailoverMemberIDs, memberID) {
		return false, fmt.Errorf("member %d is not part of failover group %d", memberID, virtualGroupID)
	}
	return s.probeMemberOnce(ctx, g, memberID), nil
}

// helpers
func indexOfInt64(arr []int64, v int64) int {
	for i, x := range arr {
		if x == v {
			return i
		}
	}
	return -1
}

func failoverContainsInt64(arr []int64, v int64) bool {
	return indexOfInt64(arr, v) >= 0
}

func triggerByPtr(id int64) *int64 {
	if id <= 0 {
		return nil
	}
	out := id
	return &out
}
