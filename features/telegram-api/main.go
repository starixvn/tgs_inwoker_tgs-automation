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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName = "telegram-api"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	config := util.GetConfig()
	serviceName := os.Getenv("SERVICE_NAME")
	fmt.Println("config.OtelCollectorUrl:", config.OpenTelemetryConfig.OtelCollectorUrl)
	ctx := context.Background()

	tp, err := opentelemetry.InitialTracerProvider(ctx, config.OpenTelemetryConfig.OtelCollectorUrl, serviceName, "0.1.0", os.Getenv("ENV"))

	if err != nil {
		log.Printf("Failed to init otel tracer:%v", err)
		return
	}
	defer tp.Shutdown(ctx)

	mp, err := opentelemetry.InitialPrometheusExporter()
	if err != nil {
		log.Printf("Failed to init otel meter: %v", err)
		return
	}
	defer mp.Shutdown(ctx)

	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.Use(otelgin.Middleware(ServiceName))
	router.Use(middleware.MetricsMiddleware(mp))
	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "10.139.0.0/16"})
	router.GET("/healthz", healthCheckHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/token", jwttoken.TokenHandler)
	router.POST("/send", middleware.AuthMiddleware(), TelegramMessageHandler())
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
