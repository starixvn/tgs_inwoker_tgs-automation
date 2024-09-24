package token

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTokenHandler(t *testing.T) {
	// 设置 Gin 的模式为测试模式
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 设置测试环境变量
	os.Setenv("API_KEY", "test_api_key")

	// 创建测试用的 JWTService 和 APIKeyAuth
	jwtService := NewJWTService()
	apiKeyAuth, err := NewAPIKeyAuth()
	if err != nil {
		t.Fatalf("Failed to initialize APIKeyAuth: %v", err)
	}

	// 注册路由
	router.GET("/token", TokenHandler(*jwtService, apiKeyAuth))

	// 构造请求
	req, _ := http.NewRequest("GET", "/token", nil)
	req.Header.Set("apikey", "test_api_key")
	w := httptest.NewRecorder()

	// 发送请求
	router.ServeHTTP(w, req)

	// 检查响应状态码
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", w.Code)
	}

	// 可以进一步检查响应体是否包含 Token
}

func TestTokenHandler_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 设置测试环境变量
	os.Setenv("API_KEY", "test_api_key")

	// 创建测试用的 JWTService 和 APIKeyAuth
	jwtService := NewJWTService()
	apiKeyAuth, err := NewAPIKeyAuth()
	if err != nil {
		t.Fatalf("Failed to initialize APIKeyAuth: %v", err)
	}

	// 注册路由
	router.GET("/token", TokenHandler(*jwtService, apiKeyAuth))

	// 构造请求，没有设置 API Key
	req, _ := http.NewRequest("GET", "/token", nil)
	w := httptest.NewRecorder()

	// 发送请求
	router.ServeHTTP(w, req)

	// 检查响应状态码
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status code 401, got %d", w.Code)
	}
}
