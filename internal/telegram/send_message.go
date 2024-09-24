package telegram

import (
	"context"
	"sync"
	"tgs-automation/internal/config"
)

// 定義 TelegramService 接口，便於測試時替換具體實現
type TelegramService interface {
	SendMessage(ctx context.Context, msg string) error
	SendMessageWithChatID(ctx context.Context, msg string, chatID string) error
}

// 定義具體的 TelegramBotService 實現
type TelegramBotService struct {
	token  string
	config config.TgsConfig
	bot    *TelegramBot
	mu     sync.Once
}

// NewTelegramService 創建 TelegramService 實例
func NewTelegramService(token string, config config.TgsConfig) TelegramService {
	return &TelegramBotService{
		token:  token,
		config: config,
	}
}

// 獲取 TelegramBot 實例，確保只初始化一次
func (s *TelegramBotService) getTelegramBotInstance() (*TelegramBot, error) {
	var err error
	s.mu.Do(func() {
		s.bot, err = newTelegramBot(s.token)
	})
	if err != nil {
		return nil, err
	}
	return s.bot, nil
}

// SendMessage 發送消息到預設的 chat ID
func (s *TelegramBotService) SendMessage(ctx context.Context, msg string) error {
	bot, err := s.getTelegramBotInstance()
	if err != nil {
		return err
	}

	bot.Bot.Debug = true
	return bot.SendMessage(s.config.Telegram.TelegramChatId, msg)
}

// SendMessageWithChatID 發送消息到指定的 chat ID
func (s *TelegramBotService) SendMessageWithChatID(ctx context.Context, msg string, chatID string) error {
	bot, err := s.getTelegramBotInstance()
	if err != nil {
		return err
	}

	bot.Bot.Debug = true
	return bot.SendMessage(chatID, msg)
}
