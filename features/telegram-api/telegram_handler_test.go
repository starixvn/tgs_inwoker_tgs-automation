package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"tgs-automation/internal/config"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/dig"
// )

// type MockTelegramBotInstance struct {
// 	ShouldFail   bool
// 	SentMessages []SendTelegramRequest
// }

// // 模拟 SendMessage 方法
// func (m *MockTelegramBotInstance) SendMessage(chatId, message string) error {
// 	// 记录发送的消息
// 	m.SentMessages = append(m.SentMessages, SendTelegramRequest{
// 		ChatId:  chatId,
// 		Message: message,
// 	})

// 	// 模拟消息发送失败
// 	if m.ShouldFail {
// 		return fmt.Errorf("failed to send message")
// 	}
// 	// 模拟消息发送成功
// 	return nil
// }

// // Test for successful message sending
// func TestTelegramMessageHandler_Success(t *testing.T) {
// 	// 创建一个模拟的 TelegramBotInstance，并确保消息发送成功
// 	mockBot := &MockTelegramBotInstance{ShouldFail: false}

// 	// 创建一个模拟的 TgsConfig，传递 Telegram Bot Token
// 	mockConfig := config.TgsConfig{
// 		Telegram: config.TelegramConfig{
// 			TelegramBotToken: "test-token",
// 		},
// 	}

// 	// 创建 DI 容器并提供 Mock 依赖
// 	container := dig.New()
// 	container.Provide(func() *MockTelegramBotInstance {
// 		return mockBot
// 	})
// 	container.Provide(func() config.TgsConfig {
// 		return mockConfig
// 	})

// 	// 设置 Gin 路由
// 	router := gin.Default()
// 	router.POST("/send", TelegramMessageHandler(mockConfig))

// 	// 模拟请求数据
// 	requestBody, _ := json.Marshal(SendTelegramRequest{
// 		ChatId:  "12345",
// 		Message: "Test message",
// 	})

// 	// 创建 HTTP 请求
// 	req, _ := http.NewRequest("POST", "/send", bytes.NewBuffer(requestBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	// 创建响应记录器
// 	w := httptest.NewRecorder()

// 	// 处理请求
// 	router.ServeHTTP(w, req)

// 	// 断言返回的状态码
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// 断言消息是否发送成功
// 	assert.Equal(t, 1, len(mockBot.SentMessages))
// 	assert.Equal(t, "12345", mockBot.SentMessages[0].ChatId)
// 	assert.Equal(t, "Test message", mockBot.SentMessages[0].Message)
// }

// // Test for message sending failure
// func TestTelegramMessageHandler_Failure(t *testing.T) {
// 	// 模拟 TelegramBotInstance 发送消息失败
// 	mockBot := &MockTelegramBotInstance{ShouldFail: true}

// 	// 创建一个模拟的 TgsConfig
// 	mockConfig := config.TgsConfig{
// 		Telegram: config.TelegramConfig{
// 			TelegramBotToken: "test-token",
// 		},
// 	}

// 	// 设置 Gin 路由
// 	router := gin.Default()
// 	router.POST("/send", TelegramMessageHandler(mockConfig))

// 	// 模拟请求数据
// 	requestBody, _ := json.Marshal(SendTelegramRequest{
// 		ChatId:  "12345",
// 		Message: "Test message",
// 	})

// 	// 创建 HTTP 请求
// 	req, _ := http.NewRequest("POST", "/send", bytes.NewBuffer(requestBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	// 创建响应记录器
// 	w := httptest.NewRecorder()

// 	// 处理请求
// 	router.ServeHTTP(w, req)

// 	// 断言返回的状态码
// 	assert.Equal(t, http.StatusInternalServerError, w.Code)

// 	// 断言消息记录
// 	assert.Equal(t, 1, len(mockBot.SentMessages))
// 	assert.Equal(t, "12345", mockBot.SentMessages[0].ChatId)
// 	assert.Equal(t, "Test message", mockBot.SentMessages[0].Message)
// }

// // Test for invalid request data
// func TestTelegramMessageHandler_InvalidData(t *testing.T) {
// 	// 创建一个模拟的 TgsConfig
// 	mockConfig := config.TgsConfig{
// 		Telegram: config.TelegramConfig{
// 			TelegramBotToken: "test-token",
// 		},
// 	}

// 	// 设置 Gin 路由
// 	router := gin.Default()
// 	router.POST("/send", TelegramMessageHandler(mockConfig))

// 	// 模拟无效请求数据（缺少 chatId）
// 	requestBody := `{"message": "Test message"}`

// 	// 创建 HTTP 请求
// 	req, _ := http.NewRequest("POST", "/send", bytes.NewBufferString(requestBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	// 创建响应记录器
// 	w := httptest.NewRecorder()

// 	// 处理请求
// 	router.ServeHTTP(w, req)

// 	// 断言返回的状态码
// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// }
