package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Task type constants
const (
	TaskTypeVerifyCode    = "verify_code"
	TaskTypePasswordReset = "password_reset"
)

// EmailOptions holds optional overrides for email sending (e.g. reseller branding, locale).
type EmailOptions struct {
	FromName string // Override SMTP From Name (e.g. reseller domain)
	SiteName string // Override site name in email content (e.g. reseller site name or domain)
	Locale   string // User locale for email content language (e.g. "en", "zh", "ja")
}

// EmailTask 邮件发送任务
type EmailTask struct {
	Email    string
	SiteName string
	TaskType string // "verify_code" or "password_reset"
	ResetURL string // Only used for password_reset task type
	Options  EmailOptions
}

// EmailQueueService 异步邮件队列服务
type EmailQueueService struct {
	emailService *EmailService
	taskChan     chan EmailTask
	wg           sync.WaitGroup
	stopChan     chan struct{}
	workers      int
}

// NewEmailQueueService 创建邮件队列服务
func NewEmailQueueService(emailService *EmailService, workers int) *EmailQueueService {
	if workers <= 0 {
		workers = 3 // 默认3个工作协程
	}

	service := &EmailQueueService{
		emailService: emailService,
		taskChan:     make(chan EmailTask, 100), // 缓冲100个任务
		stopChan:     make(chan struct{}),
		workers:      workers,
	}

	// 启动工作协程
	service.start()

	return service
}

// start 启动工作协程
func (s *EmailQueueService) start() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	log.Printf("[EmailQueue] Started %d workers", s.workers)
}

// worker 工作协程
func (s *EmailQueueService) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.taskChan:
			s.processTask(id, task)
		case <-s.stopChan:
			log.Printf("[EmailQueue] Worker %d stopping", id)
			return
		}
	}
}

// effectiveSiteName returns the site name to use in email content,
// preferring the override from EmailOptions.
func (t *EmailTask) effectiveSiteName() string {
	if t.Options.SiteName != "" {
		return t.Options.SiteName
	}
	return t.SiteName
}

// processTask 处理任务
func (s *EmailQueueService) processTask(workerID int, task EmailTask) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	siteName := task.effectiveSiteName()

	switch task.TaskType {
	case TaskTypeVerifyCode:
		if err := s.emailService.SendVerifyCode(ctx, task.Email, siteName, task.Options); err != nil {
			log.Printf("[EmailQueue] Worker %d failed to send verify code to %s: %v", workerID, task.Email, err)
		} else {
			log.Printf("[EmailQueue] Worker %d sent verify code to %s", workerID, task.Email)
		}
	case TaskTypePasswordReset:
		if err := s.emailService.SendPasswordResetEmailWithCooldown(ctx, task.Email, siteName, task.ResetURL, task.Options); err != nil {
			log.Printf("[EmailQueue] Worker %d failed to send password reset to %s: %v", workerID, task.Email, err)
		} else {
			log.Printf("[EmailQueue] Worker %d sent password reset to %s", workerID, task.Email)
		}
	default:
		log.Printf("[EmailQueue] Worker %d unknown task type: %s", workerID, task.TaskType)
	}
}

// EnqueueVerifyCode 将验证码发送任务加入队列
func (s *EmailQueueService) EnqueueVerifyCode(email, siteName string, opts EmailOptions) error {
	task := EmailTask{
		Email:    email,
		SiteName: siteName,
		TaskType: TaskTypeVerifyCode,
		Options:  opts,
	}

	select {
	case s.taskChan <- task:
		log.Printf("[EmailQueue] Enqueued verify code task for %s", email)
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// EnqueuePasswordReset 将密码重置邮件任务加入队列
func (s *EmailQueueService) EnqueuePasswordReset(email, siteName, resetURL string, opts EmailOptions) error {
	task := EmailTask{
		Email:    email,
		SiteName: siteName,
		TaskType: TaskTypePasswordReset,
		ResetURL: resetURL,
		Options:  opts,
	}

	select {
	case s.taskChan <- task:
		log.Printf("[EmailQueue] Enqueued password reset task for %s", email)
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// Stop 停止队列服务
func (s *EmailQueueService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("[EmailQueue] All workers stopped")
}
