package dnsresolver

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

// MockDNSClient 是模擬的 DNSClient 實現，用於測試
type MockDNSClient struct {
	MockResponse *dns.Msg
	MockError    error
}

// Exchange 是 Mock 的實現
func (m *MockDNSClient) Exchange(msg *dns.Msg, address string) (*dns.Msg, time.Duration, error) {
	return m.MockResponse, 0, m.MockError
}

func TestGetOriginalNameServer(t *testing.T) {
	// 創建模擬的 DNS 回應，帶有 NS 記錄
	mockResponse := new(dns.Msg)
	mockResponse.Answer = append(mockResponse.Answer, &dns.NS{
		Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 600},
		Ns:  "ns1.example.com.",
	})

	// 創建 Mock DNS 客戶端
	mockClient := &MockDNSClient{
		MockResponse: mockResponse,
		MockError:    nil,
	}

	// 創建 DNSResolver 並注入 Mock 客戶端
	resolver := NewDNSResolver(mockClient)

	// 執行查詢，並驗證結果
	ns, err := resolver.GetOriginalNameServer("example.com")
	assert.NoError(t, err)
	assert.Equal(t, "ns1.example.com.", ns)
}
