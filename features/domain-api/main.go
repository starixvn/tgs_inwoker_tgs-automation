package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "tgs-automation/features/domain-api/docs"
	"tgs-automation/internal/config"
	"tgs-automation/internal/dnsresolver"
	"tgs-automation/internal/fluentbitlogger"
	middleware "tgs-automation/internal/middleware"
	"tgs-automation/internal/nats"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/telegram"
	jwttoken "tgs-automation/internal/token"
	"tgs-automation/pkg/cloudflare"
	"tgs-automation/pkg/namecheap"

	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/swaggerFiles"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	ctx := context.Background()
	container, err := BuildContainer(ctx, os.Getenv("NODE_IP"), 24224)
	if err != nil {
		log.Fatalf("Failed to build container: %v", err)
	}
	err = container.Invoke(func(
		config config.TgsConfig,
		telemetry *opentelemetry.Telemetry,
		tokenService jwttoken.JWTService,
		apikeyAuth *jwttoken.APIKeyAuth,
		telegramBot *telegram.TelegramBot,
		cloudflareSvc cloudflare.CloudflareApi,
		namecheapSvc namecheap.NamecheapApi,
		natsSvc nats.NatsPublisherService,
		dnsResolver *dnsresolver.DNSResolver,
		logger *fluentbitlogger.Logger,
	) error {

		if err := telemetry.Initialize(ctx, ":8889"); err != nil {
			fmt.Printf("Failed to initialize telemetry: %v\n", err)
			return err
		}
		defer telemetry.Shutdown(ctx)
		router := setupRouter(
			config,
			telemetry,
			tokenService,
			apikeyAuth,
			telegramBot,
			cloudflareSvc,
			namecheapSvc,
			natsSvc,
			dnsResolver,
			logger,
		)
		server := &http.Server{
			Addr:    ":8080",
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

func setupRouter(
	config config.TgsConfig,
	telemetry *opentelemetry.Telemetry,
	tokenService jwttoken.JWTService,
	apikeyAuth *jwttoken.APIKeyAuth,
	telegramBot *telegram.TelegramBot,
	cloudflareSvc cloudflare.CloudflareApi,
	namecheapSvc namecheap.NamecheapApi,
	natsSvc nats.NatsPublisherService,
	dnsresolver *dnsresolver.DNSResolver,
	logger *fluentbitlogger.Logger) *gin.Engine {

	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})

	router.Use(otelgin.Middleware(os.Getenv("SERVICE_NAME")))
	router.Use(middleware.MetricsMiddleware(telemetry.MeterProvider))
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler(tokenService, apikeyAuth))
	router.GET("/nameserver", middleware.AuthMiddleware(), GetNameServerHandler(config, cloudflareSvc, natsSvc, telegramBot, dnsresolver, logger))
	router.PUT("/nameserver", middleware.AuthMiddleware(), UpdateNameServerHandler(config, namecheapSvc, natsSvc, telegramBot, logger))
	router.GET("/domain/price", middleware.AuthMiddleware(), GetDomainPriceHandler(config, namecheapSvc, natsSvc, telegramBot, logger))
	router.POST("/domain", middleware.AuthMiddleware(), CreateDomainHandler(config, namecheapSvc, natsSvc, telegramBot, logger))

	return router
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}
