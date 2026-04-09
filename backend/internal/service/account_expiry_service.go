package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// AccountExpiryService periodically pauses expired accounts when auto-pause is enabled.
type AccountExpiryService struct {
	accountRepo AccountRepository
	locker      *LeaderLocker
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func NewAccountExpiryService(accountRepo AccountRepository, interval time.Duration, locker *LeaderLocker) *AccountExpiryService {
	return &AccountExpiryService{
		accountRepo: accountRepo,
		locker:      locker,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (s *AccountExpiryService) Start() {
	if s == nil || s.accountRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
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
}

func (s *AccountExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *AccountExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	release, ok := s.locker.TryAcquire(ctx, "leader:account_expiry", 2*time.Minute)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	updated, err := s.accountRepo.AutoPauseExpiredAccounts(ctx, time.Now())
	if err != nil {
		log.Printf("[AccountExpiry] Auto pause expired accounts failed: %v", err)
		return
	}
	if updated > 0 {
		log.Printf("[AccountExpiry] Auto paused %d expired accounts", updated)
	}
}
