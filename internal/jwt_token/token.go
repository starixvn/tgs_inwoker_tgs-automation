package jwttoken

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

var JwtKey = []byte("secret_key")

type Claims struct {
	jwt.StandardClaims
}

// generateToken generates a JWT token
func GenerateToken() (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func TokenHandler(c *gin.Context) {
	apiKey := c.GetHeader("apikey")

	log.Println("apikey:", apiKey)

	if os.Getenv("ENV") == "local" {
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file:", err)
			return
		}
	}

	expectedApiKey := os.Getenv("API_KEY")
	log.Println("API_KEY:", expectedApiKey)

	if apiKey != expectedApiKey || apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tokenString, err := GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
