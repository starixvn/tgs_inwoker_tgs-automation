package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "tgs-automation/features/domain-api/docs"
	"tgs-automation/internal/config"
	"tgs-automation/internal/fluentbitlogger"
	middleware "tgs-automation/internal/middleware"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/telegram"
	jwttoken "tgs-automation/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/swaggerFiles"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/dig"
)

type SendTelegramRequest struct {
	ChatId      string `json:"chat_id"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
}

func BuildContainer(nodeip string, port int) (*dig.Container, error) {
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

	container.Provide(func() (*fluentbitlogger.Logger, error) {
		return fluentbitlogger.NewLogger(nodeip, port)
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

func main() {
	gin.SetMode(gin.ReleaseMode)
	container, err := BuildContainer(os.Getenv("NODE_IP"), 24224)
	if err != nil {
		log.Fatalf("Failed to build container: %v", err)
	}
	err = container.Invoke(func(
		config config.TgsConfig,
		telemetry *opentelemetry.Telemetry,
		tokenService jwttoken.JWTService,
		apikeyAuth *jwttoken.APIKeyAuth,
		telegramBot *telegram.TelegramBot,
	) error {
		ctx := context.Background()
		if err := telemetry.Initialize(ctx, ":8889"); err != nil {
			fmt.Printf("Failed to initialize telemetry: %v\n", err)
			return err
		}
		defer telemetry.Shutdown(ctx)
		router := setupRouter(telemetry, tokenService, apikeyAuth, telegramBot)
		server := &http.Server{
			Addr:    ":8081",
			Handler: router,
		}
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe(): %v", err)
			}
		}()
		select {}
	})

	if err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

func setupRouter(
	telemetry *opentelemetry.Telemetry,
	tokenService jwttoken.JWTService,
	apikeyAuth *jwttoken.APIKeyAuth,
	telegramBot *telegram.TelegramBot) *gin.Engine {
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})

	router.Use(otelgin.Middleware(os.Getenv("SERVICE_NAME")))
	router.Use(middleware.MetricsMiddleware(telemetry.MeterProvider))
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler(tokenService, apikeyAuth))
	router.POST("/send", middleware.AuthMiddleware(), TelegramMessageHandler(telegramBot))
	return router
}
