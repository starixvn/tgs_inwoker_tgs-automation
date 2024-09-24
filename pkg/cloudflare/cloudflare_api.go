package cloudflare

import (
	"tgs-automation/internal/config"
)

// mockgen -destination pkg/cloudflare/mock_cloudflare.go -package=cloudflare tgs-automation/pkg/Cloudflare CloudflareApi
// go generate mockgen -destination ../mocks/mock_cloudflare.go -package=cloudflare tgs-automation/pkg/Cloudflare CloudflareApi
type CloudflareApi interface {
	CreateDNS(domain string) error
	DeleteDNS(domain string) error
	GetDnsInfo(domain string) error
	GetTargetNameServers(domain string) (string, error)
}

type CloudflareService struct {
	Config config.TgsConfig
}

func NewCloudflare(config config.TgsConfig) CloudflareApi {
	return &CloudflareService{
		Config: config,
	}
}
