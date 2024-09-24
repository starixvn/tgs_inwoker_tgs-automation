package main

import (
	"log"
	"net/http"
	_ "tgs-automation/features/domain-api/docs"
	"tgs-automation/internal/telegram"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

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
func TelegramMessageHandler(telegramBot *telegram.TelegramBot) gin.HandlerFunc {
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

		err := telegramBot.SendMessage(request.ChatId, request.Message)

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
