package main

import (
	"fmt"
	"net/http"
	"strings"
	"tgs-automation/internal/config"
	"tgs-automation/internal/dnsresolver"
	"tgs-automation/internal/log"
	"tgs-automation/internal/nats"
	"tgs-automation/internal/telegram"
	"tgs-automation/pkg/cloudflare"
	"tgs-automation/pkg/namecheap"

	"github.com/gin-gonic/gin"
)

type GetNameServerRequest struct {
	Domain string `form:"domain" json:"domain" binding:"required"`
	ChatId string `form:"chatid" json:"chatid" binding:"required"`
}

type UpdateNameServerRequest struct {
	Domain      string `form:"domain" json:"domain" binding:"required"`
	ChatId      string `form:"chatid" json:"chatid" binding:"required"`
	NameServers string `form:"nameservers" json:"nameservers" binding:"required"`
}

// UpdateNameServer changes the name server information
// @Summary Update name server information
// @Tags NameServer
// @Description Update the name server information for a given domain
// @Accept json
// @Produce json
// @Param updateNameServerRequest body UpdateNameServerRequest true "Update name server request"
// @Success 200 {object} ApiResponse "Success"
// @Failure 400 {object} ApiResponse "Bad Request"
// @Security ApiKeyAuth
// @Router /nameservers [put]
func UpdateNameServerHandler(
	config config.TgsConfig,
	svc namecheap.NamecheapApi,
	natsSvc nats.NatsPublisherService,
	telegramBot *telegram.TelegramBot,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		printAllFields(c)
		var request UpdateNameServerRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update nameserver request data", "details": err.Error()})
			return
		}

		response, err := svc.UpdateNameServer(request.Domain, request.NameServers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update nameserver", "details": err.Error()})
			return
		}

		var message string
		if strings.Contains(response.Status, "OK") {
			message = fmt.Sprintf("修改nameserver: %s 成功\nNameServers: %s", request.Domain, request.NameServers)
		} else {
			message = fmt.Sprintf("修改nameserver: %s 失敗\n原因: %s\nNameServers: %s", request.Domain, response.Errors.Error, request.NameServers)
		}

		err = telegramBot.SendMessageWithChatIdAndContext(c.Request.Context(), request.ChatId, message)
		if err != nil {
			fmt.Println("Failed to send Telegram message:", err)
		}

		c.JSON(http.StatusOK, gin.H{"data": message})
	}
}

// GetNameServer retrieves the name server information
// @Summary Retrieve name server information
// @Tags NameServer
// @Description Retrieve the name server information for a given domain
// @Accept json
// @Produce json
// @Param domain query string true "Domain name"
// @Param chatid query string true "Chat ID"
// @Success 200 {object} ApiResponse "Success"
// @Failure 400 {object} ApiResponse "Bad Request"
// @Security ApiKeyAuth
// @Router /nameservers [get]
func GetNameServerHandler(
	config config.TgsConfig,
	cloudflareApi cloudflare.CloudflareApi,
	natsSvc nats.NatsPublisherService,
	telegramBot *telegram.TelegramBot,
	resolver *dnsresolver.DNSResolver,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GetNameServerRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			return
		}

		log.LogInfo(fmt.Sprintf("Request data: %+v", request))
		targetNameServer, err := cloudflareApi.GetTargetNameServers(request.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get target nameserver", "details": err.Error()})
			return
		}

		originNameServer, err := resolver.GetOriginalNameServer(request.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get original nameserver", "details": err.Error()})
			return
		}

		message := fmt.Sprintf("Domain: %s\nNameServers: %s\nOriginal Nameservers: %s", request.Domain, targetNameServer, originNameServer)

		//log.LogInfo(fmt.Sprintf("TelegramApi", config.ApiUrl.TelegramApi))

		err = telegramBot.SendMessageWithChatIdAndContext(c.Request.Context(), request.ChatId, message)

		if err != nil {
			fmt.Println("Failed to send Telegram message:", err)
		}
		c.JSON(http.StatusOK, gin.H{"data": message})
	}
}
