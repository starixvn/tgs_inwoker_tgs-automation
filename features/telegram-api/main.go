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
	"tgs-automation/pkg/telegram"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/swaggerFiles"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {

	serviceName := "telegram-api"
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	config := util.GetConfig()
	fmt.Println("config.OtelCollectorUrl:", config.OpenTelemetryConfig.OtelCollectorUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cleanup, err := opentelemetry.InitTracer(ctx, config.OpenTelemetryConfig.OtelCollectorUrl, serviceName, "0.1.0", "prod")

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

	router.Use(otelgin.Middleware(serviceName))

	if err != nil {
		fmt.Println("cannot get nats svc.")
	}
	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler)
	router.POST("/send", middleware.AuthMiddleware(), NatsHandler())
	err = server.ListenAndServe()

	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

type SendTelegramRequest struct {
	ChatId      string `json:"chat_id"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
}

// NatsHandler
// @Summary NatsHandler
// @Tags message
// @Description NatsHandler
// @Accept json
// @Produce json
// @Param message query string true "message"
// @Param chatid query string true "chatid"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} map[string]interface{} "error"
// @Router /send [post]
func NatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request SendTelegramRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		config := util.GetConfig()

		bot, err := telegram.GetTelegramBotInstance(config.Telegram.TelegramBotToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error getting Telegram bot instance": err.Error()})
			return
		}

		err = bot.SendMessage(request.ChatId, request.Message)

		if err != nil {
			log.Printf("Error sending Telegram message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error sending Telegram message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Sending Telegram message successfully."})
	}
}
