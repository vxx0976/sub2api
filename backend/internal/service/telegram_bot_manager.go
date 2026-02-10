package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBotManager manages the lifecycle of per-reseller Telegram bots.
type TelegramBotManager struct {
	settingRepo ResellerSettingRepository
	apiKeyRepo  APIKeyRepository
	resellerSvc *ResellerService
	mu          sync.RWMutex
	bots        map[int64]*botInstance // resellerID -> running bot
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

type botInstance struct {
	resellerID int64
	bot        *tgbotapi.BotAPI
	cancel     context.CancelFunc
	features   map[string]bool
}

// NewTelegramBotManager creates a new TelegramBotManager.
func NewTelegramBotManager(
	settingRepo ResellerSettingRepository,
	apiKeyRepo APIKeyRepository,
	resellerSvc *ResellerService,
) *TelegramBotManager {
	return &TelegramBotManager{
		settingRepo: settingRepo,
		apiKeyRepo:  apiKeyRepo,
		resellerSvc: resellerSvc,
		bots:        make(map[int64]*botInstance),
		stopCh:      make(chan struct{}),
	}
}

// Start scans all resellers with configured bot tokens and starts their bots.
func (m *TelegramBotManager) Start() {
	if m == nil {
		return
	}

	ctx := context.Background()
	tokens, err := m.settingRepo.ListByKey(ctx, ResellerSettingTgBotToken)
	if err != nil {
		log.Printf("[TelegramBot] Failed to list bot tokens on startup: %v", err)
		return
	}

	for resellerID, token := range tokens {
		if token == "" {
			continue
		}
		// Load features
		features := m.loadFeatures(ctx, resellerID)
		if err := m.startBot(resellerID, token, features); err != nil {
			log.Printf("[TelegramBot] Failed to start bot for reseller %d: %v", resellerID, err)
		}
	}

	log.Printf("[TelegramBot] Started %d bots", len(m.bots))
}

// Stop gracefully stops all running bots.
func (m *TelegramBotManager) Stop() {
	if m == nil {
		return
	}
	close(m.stopCh)

	m.mu.Lock()
	for _, inst := range m.bots {
		inst.cancel()
		inst.bot.StopReceivingUpdates()
	}
	m.mu.Unlock()

	m.wg.Wait()
	log.Printf("[TelegramBot] All bots stopped")
}

// OnSettingsUpdated is called when a reseller's settings change.
// It restarts the bot if the token or features changed.
func (m *TelegramBotManager) OnSettingsUpdated(resellerID int64, settings map[string]string) {
	if m == nil {
		return
	}

	token := settings[ResellerSettingTgBotToken]
	features := parseTgFeatures(settings[ResellerSettingTgFeatures])

	m.mu.RLock()
	existing := m.bots[resellerID]
	m.mu.RUnlock()

	if token == "" {
		// Token removed → stop bot
		if existing != nil {
			m.stopBot(resellerID)
		}
		return
	}

	// Check if restart needed
	needRestart := existing == nil
	if existing != nil {
		// Token changed?
		// We can't read back the token from tgbotapi, so always restart if settings updated
		needRestart = true
	}

	if needRestart {
		if existing != nil {
			m.stopBot(resellerID)
		}
		if err := m.startBot(resellerID, token, features); err != nil {
			log.Printf("[TelegramBot] Failed to restart bot for reseller %d: %v", resellerID, err)
		}
	}
}

// SendNotification sends a message to a specific chat via a reseller's bot.
func (m *TelegramBotManager) SendNotification(resellerID int64, chatID int64, text string) error {
	m.mu.RLock()
	inst := m.bots[resellerID]
	m.mu.RUnlock()

	if inst == nil {
		return fmt.Errorf("no bot running for reseller %d", resellerID)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := inst.bot.Send(msg)
	return err
}

func (m *TelegramBotManager) startBot(resellerID int64, token string, features map[string]bool) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	inst := &botInstance{
		resellerID: resellerID,
		bot:        bot,
		cancel:     cancel,
		features:   features,
	}

	m.mu.Lock()
	m.bots[resellerID] = inst
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runBot(ctx, inst)
	}()

	log.Printf("[TelegramBot] Started bot for reseller %d (@%s)", resellerID, bot.Self.UserName)
	return nil
}

func (m *TelegramBotManager) stopBot(resellerID int64) {
	m.mu.Lock()
	inst := m.bots[resellerID]
	if inst != nil {
		inst.cancel()
		inst.bot.StopReceivingUpdates()
		delete(m.bots, resellerID)
	}
	m.mu.Unlock()

	if inst != nil {
		log.Printf("[TelegramBot] Stopped bot for reseller %d", resellerID)
	}
}

func (m *TelegramBotManager) runBot(ctx context.Context, inst *botInstance) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := inst.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}
			m.handleMessage(inst, update.Message)
		}
	}
}

func (m *TelegramBotManager) loadFeatures(ctx context.Context, resellerID int64) map[string]bool {
	raw, err := m.settingRepo.Get(ctx, resellerID, ResellerSettingTgFeatures)
	if err != nil || raw == "" {
		return parseTgFeatures(DefaultTgFeatures)
	}
	return parseTgFeatures(raw)
}

// getResellerChatID returns the chat ID bound to the reseller (admin mode).
func (m *TelegramBotManager) getResellerChatID(resellerID int64) int64 {
	val, err := m.settingRepo.Get(context.Background(), resellerID, ResellerSettingTgChatID)
	if err != nil || val == "" {
		return 0
	}
	var chatID int64
	fmt.Sscanf(val, "%d", &chatID)
	return chatID
}

// parseTgFeatures parses a comma-separated feature string into a map.
func parseTgFeatures(s string) map[string]bool {
	features := make(map[string]bool)
	if s == "" {
		// All enabled by default
		for _, f := range strings.Split(DefaultTgFeatures, ",") {
			features[strings.TrimSpace(f)] = true
		}
		return features
	}
	for _, f := range strings.Split(s, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			features[f] = true
		}
	}
	return features
}

// featureEnabled checks if a feature is enabled for the bot instance.
func featureEnabled(inst *botInstance, feature string) bool {
	if inst.features == nil {
		return true // default: all enabled
	}
	return inst.features[feature]
}

// sendReply sends a text reply to the given message.
func sendReply(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("[TelegramBot] Failed to send message to chat %d: %v", chatID, err)
	}
}

// ProvideTelegramBotManager creates and starts TelegramBotManager.
func ProvideTelegramBotManager(
	settingRepo ResellerSettingRepository,
	apiKeyRepo APIKeyRepository,
	resellerSvc *ResellerService,
) *TelegramBotManager {
	mgr := NewTelegramBotManager(settingRepo, apiKeyRepo, resellerSvc)
	// Delay startup slightly to let other services initialize
	go func() {
		time.Sleep(3 * time.Second)
		mgr.Start()
	}()
	return mgr
}
