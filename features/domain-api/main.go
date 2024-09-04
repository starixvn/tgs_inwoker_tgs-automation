package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "tgs-automation/features/domain-api/docs"
	jwttoken "tgs-automation/internal/jwt_token"
	middleware "tgs-automation/internal/middleware"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/util"
	"tgs-automation/pkg/cloudflare"
	"tgs-automation/pkg/namecheap"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/swaggerFiles"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	config := util.GetConfig()
	fmt.Println("config.OtelCollectorUrl:", config.OpenTelemetryConfig.OtelCollectorUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cleanup, err := opentelemetry.InitTracer(ctx, config.OpenTelemetryConfig.OtelCollectorUrl, "domain-api", "0.1.0", "prod")

	if err != nil {
		log.Printf("Failed to init otel tracer:%v", err)
		return
	}
	defer func() {
		cancel()
		if err := cleanup(ctx); err != nil {
			log.Printf("Failed to clueanup otel tracer: %v", err)
		}
	}()

	router.Use(otelgin.Middleware("domain-api"))
	natsSvc, err := util.NewNatsPublisher(config.NatsUrl)

	defer func() {
		natsSvc.Close()
	}()

	if err != nil {
		fmt.Println("cannot get nats svc.")
	}

	namecheapSvc := namecheap.New(config)
	cloudflareSvc := cloudflare.NewClloudflare(config)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler)
	router.GET("/nameserver", middleware.AuthMiddleware(), GetNameServerHandler(config, cloudflareSvc, natsSvc))
	router.PUT("/nameserver", middleware.AuthMiddleware(), UpdateNameServerHandler(config, namecheapSvc, natsSvc))
	router.GET("/domain/price", middleware.AuthMiddleware(), GetDomainPriceHandler(config, namecheapSvc, natsSvc))
	router.POST("/domain", middleware.AuthMiddleware(), CreateDomainHandler(config, namecheapSvc, natsSvc))
	err = server.ListenAndServe()

	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}
