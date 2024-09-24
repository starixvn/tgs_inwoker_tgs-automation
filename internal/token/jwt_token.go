package token

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	tgshttp "tgs-automation/internal/http"
	"tgs-automation/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type TokenService interface {
	GenerateToken() (string, error)
	GetToken(ctx context.Context, url, apikey string) (string, error)
}

type JWTService struct {
	JwtKey []byte
}

var JwtKey = []byte("secret_key")

func NewJWTService() *JWTService {
	return &JWTService{
		JwtKey: []byte(JwtKey),
	}
}

type Claims struct {
	jwt.StandardClaims
}

func (j *JWTService) GenerateToken() (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.JwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (j *JWTService) GetToken(ctx context.Context, url, apikey string) (string, error) {
	resp, err := tgshttp.SendHTTPRequest(ctx, "GET", url, "", apikey, nil)
	if err != nil {
		log.LogError("SendHTTPRequest Error:" + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	var result TokenResponse
	if err := tgshttp.ParseJSONResponse(resp.Body, &result); err != nil {
		log.LogError("Parse Error:" + err.Error())
		return "", err
	}

	return result.Token, nil
}

type APIKeyAuth struct {
	ExpectedAPIKey string
}

func NewAPIKeyAuth() (*APIKeyAuth, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return nil, err
	}

	expectedApiKey := os.Getenv("API_KEY")
	return &APIKeyAuth{
		ExpectedAPIKey: expectedApiKey,
	}, nil
}

func (a *APIKeyAuth) IsAuthorized(apiKey string) bool {
	return apiKey != "" && apiKey == a.ExpectedAPIKey
}

func TokenHandler(tokenService TokenService, apiKeyAuth *APIKeyAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("apikey")

		log.LogInfo("apikey:" + apiKey)
		log.LogInfo("Expected API_KEY:" + apiKeyAuth.ExpectedAPIKey)

		if !apiKeyAuth.IsAuthorized(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString, err := tokenService.GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}
