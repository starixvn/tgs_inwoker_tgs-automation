package main

import (
	"fmt"
	"net/http"
	"strings"
	"tgs-automation/internal/log"
	"tgs-automation/internal/util"
	"tgs-automation/pkg/cloudflare"
	"tgs-automation/pkg/namecheap"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
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
	svc namecheap.NamecheapApi,
	natsSvc util.NatsPublisherService,
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

		// traceCtx := trace.SpanContextFromContext(ctx)
		// header := make(nats.Header)
		// header.Set("traceparent", traceCtx.TraceState().String())
		// propagator := propagation.TraceContext{}
		// propagator.Inject(ctx, propagation.HeaderCarrier(header))

		_ = natsSvc.PublishWithContext(c.Request.Context(), "domain-api", request.ChatId, message)

		// span.AddEvent("end to send message to telegram", trace.WithTimestamp(time.Now()))
		// span.SetStatus(codes.Ok, "")
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
	cloudflareApi cloudflare.CloudflareApi,
	natsSvc util.NatsPublisherService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request GetNameServerRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
			//recordError(span, "invalid request data", err)
			return
		}

		log.LogInfo(fmt.Sprintf("Request data: %+v", request))
		targetNameServer, err := cloudflareApi.GetTargetNameServers(request.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get target nameserver", "details": err.Error()})
			return
		}

		originNameServer, err := getOriginalNameServer(request.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get original nameserver", "details": err.Error()})
			return
		}

		message := fmt.Sprintf("Domain: %s\nNameServers: %s\nOriginal Nameservers: %s", request.Domain, targetNameServer, originNameServer)

		err = natsSvc.Publish(request.ChatId, message)
		if err != nil {
			fmt.Println("Failed to send Telegram message:", err)
		}
		c.JSON(http.StatusOK, gin.H{"data": message})
	}
}

// func recordError(span trace.Span, message string, err error) {
// 	span.RecordError(fmt.Errorf("%v: %w", message, err))
// 	span.SetStatus(codes.Error, message)
// }

func getOriginalNameServer(domain string) (string, error) {
	var nsRecords []string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.RecursionDesired = true

	c := new(dns.Client)
	in, _, err := c.Exchange(m, "8.8.8.8:53") // Using Google's public DNS server
	if err != nil {
		return "", err
	}
	fmt.Println("in.Answer=", in.Answer)
	if len(in.Answer) == 0 {
		//return "", fmt.Errorf("no NS records found for domain %s", domain)
		return "", nil
	}

	for _, ans := range in.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nsRecords = append(nsRecords, ns.Ns)
		}
	}

	return strings.Join(nsRecords, " "), nil
}
