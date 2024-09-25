package main

import (
	"context"
	"fmt"
	"tgs-automation/internal/config"
	"tgs-automation/internal/dnsresolver"
	"tgs-automation/internal/fluentbitlogger"
	"tgs-automation/internal/nats"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/telegram"
	jwttoken "tgs-automation/internal/token"
	"tgs-automation/pkg/cloudflare"
	"tgs-automation/pkg/namecheap"

	"go.uber.org/dig"
)

func BuildContainer(ctx context.Context, nodeip string, port int, serivceName string) (*dig.Container, error) {
	container := dig.New()
	if err := container.Provide(config.NewConfigManager); err != nil {
		return nil, fmt.Errorf("failed to provide ConfigManager: %w", err)
	}
	if err := container.Provide(func(configMgr *config.ConfigManager) (config.TgsConfig, error) {
		ctx := context.Background()
		return configMgr.GetConfig(ctx)
	}); err != nil {
		return nil, fmt.Errorf("failed to provide TgsConfig: %w", err)
	}
	if err := container.Provide(opentelemetry.NewTelemetry); err != nil {
		return nil, fmt.Errorf("failed to provide Telemetry: %w", err)
	}
	if err := container.Provide(func() jwttoken.JWTService {
		return *jwttoken.NewJWTService()
	}); err != nil {
		return nil, fmt.Errorf("failed to provide JWTService: %w", err)
	}
	if err := container.Provide(jwttoken.NewAPIKeyAuth); err != nil {
		return nil, fmt.Errorf("failed to provide APIKeyAuth: %w", err)
	}
	if err := container.Provide(ProvideTelegramBot); err != nil {
		return nil, fmt.Errorf("failed to provide TelegramBotInstance: %w", err)
	}

	if err := container.Provide(ProvideNamecheapService); err != nil {
		return nil, fmt.Errorf("failed to provide TelegramBotInstance: %w", err)
	}

	if err := container.Provide(ProvideCloudflareService); err != nil {
		return nil, fmt.Errorf("failed to provide TelegramBotInstance: %w", err)
	}

	if err := container.Provide(ProvideNatsPublisher); err != nil {
		return nil, fmt.Errorf("failed to provide TelegramBotInstance: %w", err)
	}

	if err := container.Provide(func() dnsresolver.DNSClient {
		return &dnsresolver.RealDNSClient{}
	}); err != nil {
		return nil, fmt.Errorf("failed to provide RealDNSClient: %w", err)
	}

	if err := container.Provide(dnsresolver.NewDNSResolver); err != nil {
		return nil, fmt.Errorf("failed to provide DNSResolver: %w", err)
	}

	container.Provide(func() (*fluentbitlogger.Logger, error) {
		return fluentbitlogger.NewLogger(nodeip, port, serivceName)
	})
	return container, nil
}

func ProvideTelegramBot(config config.TgsConfig) (*telegram.TelegramBot, error) {
	bot, err := telegram.GetTelegramBotInstance(config.Telegram.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf(`failed to initialize Telegram bot: %w`, err)
	}
	return bot, nil
}

// 提供 ConfigManager
func ProvideConfigManager() (*config.ConfigManager, error) {
	return config.NewConfigManager(), nil
}

// 提供配置
func ProvideConfig(configManager *config.ConfigManager) (config.TgsConfig, error) {
	ctx := context.Background()
	return configManager.GetConfig(ctx)
}

// 提供 NatsPublisher
func ProvideNatsPublisher(cfg config.TgsConfig) (nats.NatsPublisherService, error) {
	natsSvc, err := nats.NewNatsPublisher(cfg.NatsUrl)
	if err != nil {
		return nil, err
	}
	return natsSvc, nil
}

// 提供 Namecheap 服務
func ProvideNamecheapService(cfg config.TgsConfig) namecheap.NamecheapApi {
	return namecheap.New(cfg)
}

// 提供 Cloudflare 服務
func ProvideCloudflareService(cfg config.TgsConfig) cloudflare.CloudflareApi {
	return cloudflare.NewCloudflare(cfg)
}
