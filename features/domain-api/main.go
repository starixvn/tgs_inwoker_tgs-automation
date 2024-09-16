package main

import (
	"context"
	"fmt"
	"net/http"
	_ "tgs-automation/features/domain-api/docs"
	jwttoken "tgs-automation/internal/jwt_token"
	middleware "tgs-automation/internal/middleware"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/util"
	"tgs-automation/pkg/cloudflare"
	"tgs-automation/pkg/namecheap"

	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/swaggerFiles"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	config := util.GetConfig()
	ctx := context.Background()

	tp, mp := opentelemetry.InitialOpenTelemetry(ctx, config)
	defer func() {
		tp.Shutdown(ctx)
		mp.Shutdown(ctx)
	}()

	natsSvc, err := util.NewNatsPublisher(config.NatsUrl)
	defer func() {
		natsSvc.Close()
	}()

	if err != nil {
		fmt.Println("cannot get nats svc.")
	}

	namecheapSvc := namecheap.New(config)
	cloudflareSvc := cloudflare.NewCloudflare(config)
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler)
	router.GET("/nameserver", middleware.AuthMiddleware(), GetNameServerHandler(config, cloudflareSvc, natsSvc))
	router.PUT("/nameserver", middleware.AuthMiddleware(), UpdateNameServerHandler(config, namecheapSvc, natsSvc))
	router.GET("/domain/price", middleware.AuthMiddleware(), GetDomainPriceHandler(config, namecheapSvc, natsSvc))
	router.POST("/domain", middleware.AuthMiddleware(), CreateDomainHandler(config, namecheapSvc, natsSvc))
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Failed to start server:", err)
			panic(err)
		}
	}()
	select {}
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}
