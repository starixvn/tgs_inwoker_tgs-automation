package dnsresolver

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSClient 是一個接口，封裝了 DNS 客戶端的 Exchange 方法
type DNSClient interface {
	Exchange(m *dns.Msg, address string) (*dns.Msg, time.Duration, error)
}

// RealDNSClient 是真實的 DNSClient 實現，使用 miekg/dns
type RealDNSClient struct{}

// Exchange 執行 DNS 查詢
func (c *RealDNSClient) Exchange(m *dns.Msg, address string) (*dns.Msg, time.Duration, error) {
	client := new(dns.Client)
	return client.Exchange(m, address)
}

// DNSResolver 是用來解析域名的對象
type DNSResolver struct {
	Client DNSClient
}

// NewDNSResolver 創建一個 DNSResolver 並注入 DNS 客戶端
func NewDNSResolver(client DNSClient) *DNSResolver {
	return &DNSResolver{
		Client: client,
	}
}

// GetOriginalNameServer 查詢給定域名的原始 NS 記錄
func (r *DNSResolver) GetOriginalNameServer(domain string) (string, error) {
	var nsRecords []string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.RecursionDesired = true

	in, _, err := r.Client.Exchange(m, "8.8.8.8:53") // 使用 Google 公共 DNS 伺服器
	if err != nil {
		return "", err
	}

	fmt.Println("in.Answer=", in.Answer)
	if len(in.Answer) == 0 {
		return "", nil
	}

	for _, ans := range in.Answer {
		if ns, ok := ans.(*dns.NS); ok {
			nsRecords = append(nsRecords, ns.Ns)
		}
	}

	return strings.Join(nsRecords, " "), nil
}
