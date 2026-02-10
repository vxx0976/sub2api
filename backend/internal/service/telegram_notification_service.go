package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// TelegramNotificationService periodically checks quota/balance and sends alerts.
type TelegramNotificationService struct {
	botManager *TelegramBotManager
	interval   time.Duration
	notified   map[string]time.Time // dedup key -> last notification time
	notifiedMu sync.Mutex
	stopCh     chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

// NewTelegramNotificationService creates a new notification service.
func NewTelegramNotificationService(
	botManager *TelegramBotManager,
	interval time.Duration,
) *TelegramNotificationService {
	return &TelegramNotificationService{
		botManager: botManager,
		interval:   interval,
		notified:   make(map[string]time.Time),
		stopCh:     make(chan struct{}),
	}
}

// Start begins the periodic notification check loop.
func (s *TelegramNotificationService) Start() {
	if s == nil || s.botManager == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		// Wait a bit before first run
		select {
		case <-time.After(30 * time.Second):
		case <-s.stopCh:
			return
		}

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
	log.Printf("[TelegramNotify] Started (interval=%s)", s.interval)
}

// Stop stops the notification service.
func (s *TelegramNotificationService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
	log.Printf("[TelegramNotify] Stopped")
}

func (s *TelegramNotificationService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Clean up old dedup entries (> 24h)
	s.cleanupNotified()

	resellerIDs := s.botManager.ListRunningBotResellerIDs()
	for _, resellerID := range resellerIDs {
		s.checkReseller(ctx, resellerID)
	}
}

func (s *TelegramNotificationService) checkReseller(ctx context.Context, resellerID int64) {
	adminNotify := s.botManager.IsFeatureEnabled(resellerID, "admin_notify")
	userNotify := s.botManager.IsFeatureEnabled(resellerID, "user_notify")

	if !adminNotify && !userNotify {
		return
	}

	adminChatID := s.botManager.getResellerChatID(resellerID)

	// Check API key quotas
	keys, err := s.botManager.ListKeysForNotification(ctx, resellerID)
	if err != nil {
		log.Printf("[TelegramNotify] Error listing keys for reseller %d: %v", resellerID, err)
		return
	}

	for _, key := range keys {
		if key.Quota <= 0 {
			continue // unlimited, skip
		}
		pct := key.QuotaUsed / key.Quota * 100

		// Quota exhausted (>=100%)
		if pct >= 100 {
			if adminNotify && adminChatID != 0 {
				dedupKey := fmt.Sprintf("admin_quota_100_%d_%d", resellerID, key.ID)
				if s.shouldNotify(dedupKey) {
					_ = s.botManager.SendNotification(resellerID, adminChatID,
						fmt.Sprintf("🚨 *配额告警*\n\n密钥 `%s` (#%d) 配额已用完\n已用: $%.2f / $%.2f",
							maskKey(key.Key), key.ID, key.QuotaUsed, key.Quota))
				}
			}
			if userNotify && key.TgChatID != nil {
				dedupKey := fmt.Sprintf("user_quota_100_%d", key.ID)
				if s.shouldNotify(dedupKey) {
					_ = s.botManager.SendNotification(resellerID, *key.TgChatID,
						"🚨 *配额告警*\n\n您的密钥配额已用完，请联系管理员续费")
				}
			}
		} else if pct >= 80 {
			// Quota warning (>=80%)
			if adminNotify && adminChatID != 0 {
				dedupKey := fmt.Sprintf("admin_quota_80_%d_%d", resellerID, key.ID)
				if s.shouldNotify(dedupKey) {
					_ = s.botManager.SendNotification(resellerID, adminChatID,
						fmt.Sprintf("⚠️ *配额提醒*\n\n密钥 `%s` (#%d) 配额即将用完\n已用: $%.2f / $%.2f (%.1f%%)",
							maskKey(key.Key), key.ID, key.QuotaUsed, key.Quota, pct))
				}
			}
			if userNotify && key.TgChatID != nil {
				dedupKey := fmt.Sprintf("user_quota_80_%d", key.ID)
				if s.shouldNotify(dedupKey) {
					_ = s.botManager.SendNotification(resellerID, *key.TgChatID,
						fmt.Sprintf("⚠️ *配额提醒*\n\n您的密钥配额已用 %.1f%%，请注意续费", pct))
				}
			}
		}
	}

	// Check reseller balance
	if adminNotify && adminChatID != 0 {
		balance, err := s.botManager.GetResellerBalance(ctx, resellerID)
		if err == nil && balance < 10.0 {
			dedupKey := fmt.Sprintf("admin_balance_%d", resellerID)
			if s.shouldNotify(dedupKey) {
				_ = s.botManager.SendNotification(resellerID, adminChatID,
					fmt.Sprintf("⚠️ *余额不足*\n\n您的余额仅剩 $%.2f，请及时充值", balance))
			}
		}
	}
}

// shouldNotify checks if a notification with the given dedup key should be sent.
// Returns true if 24h has passed since last notification for this key.
func (s *TelegramNotificationService) shouldNotify(key string) bool {
	s.notifiedMu.Lock()
	defer s.notifiedMu.Unlock()

	if last, ok := s.notified[key]; ok {
		if time.Since(last) < 24*time.Hour {
			return false
		}
	}
	s.notified[key] = time.Now()
	return true
}

func (s *TelegramNotificationService) cleanupNotified() {
	s.notifiedMu.Lock()
	defer s.notifiedMu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for key, t := range s.notified {
		if t.Before(cutoff) {
			delete(s.notified, key)
		}
	}
}

// ProvideTelegramNotificationService creates and starts the notification service.
func ProvideTelegramNotificationService(
	botManager *TelegramBotManager,
) *TelegramNotificationService {
	svc := NewTelegramNotificationService(botManager, 5*time.Minute)
	svc.Start()
	return svc
}
