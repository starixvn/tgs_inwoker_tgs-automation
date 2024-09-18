package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "tgs-automation/features/domain-api/docs"
	jwttoken "tgs-automation/internal/jwt_token"
	middleware "tgs-automation/internal/middleware"
	"tgs-automation/internal/opentelemetry"
	"tgs-automation/internal/util"
	"tgs-automation/pkg/telegram"

	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/swaggerFiles"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	config := util.GetConfig()
	fmt.Println("config.OtelCollectorUrl:", config.OpenTelemetryConfig.OtelCollectorUrl)
	ctx := context.Background()

	tp, mp := opentelemetry.InitialOpenTelemetry(ctx, config, ":8889")

	defer func() {
		tp.Shutdown(ctx)
		mp.Shutdown(ctx)
	}()

	router := gin.Default()
	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})
	router.Use(otelgin.Middleware(os.Getenv("SERVICE_NAME")))
	router.Use(middleware.MetricsMiddleware(mp))
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler)
	router.POST("/send", middleware.AuthMiddleware(), TelegramMessageHandler())
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	select {}
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
func TelegramMessageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		defer span.End() // 确保在函数结束时结束 span

		//获取 TraceID
		traceID := span.SpanContext().TraceID().String()

		log.Println("traceId:" + traceID)

		var request SendTelegramRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		//在 span 中记录请求的 chatId 和 message
		span.SetAttributes(
			attribute.String("telegram.chat_id", request.ChatId),
			attribute.String("telegram.message", request.Message),
			attribute.String("trace_id", traceID), // 记录 TraceID
		)

		config := util.GetConfig()

		bot, err := telegram.GetTelegramBotInstance(config.Telegram.TelegramBotToken)
		if err != nil {
			span.RecordError(err) // 在 span 中记录错误
			span.SetStatus(codes.Error, "Error getting Telegram bot instance")
			c.JSON(http.StatusInternalServerError, gin.H{"error getting Telegram bot instance": err.Error()})
			return
		}

		err = bot.SendMessage(request.ChatId, request.Message)

		if err != nil {
			log.Printf("Error sending Telegram message: %v", err)
			span.RecordError(err) // 在 span 中记录错误
			span.SetStatus(codes.Error, "Error sending Telegram message")
			c.JSON(http.StatusInternalServerError, gin.H{"Error sending Telegram message": err.Error()})
			return
		}

		span.SetStatus(codes.Ok, "Message sent successfully")
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Sending Telegram message successfully."})
	}
}
