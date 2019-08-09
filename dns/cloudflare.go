package dns

import (
	"net"
	"net/http"
	"time"
)

// CloudflareDNSRecord represents a DNS record API object on cloudflare
// This structure will be serialized and send to the cloudflare api
type CloudflareDNSRecord struct {
	dnsType  string `json:"type"`
	name     string `json:"name"`
	content  net.IP `json:"content"`
	ttl      uint16 `json:"ttl"`
	priority uint16 `json:"priority"`
	proxied  bool   `json:"proxied"`
}

func NewFromDNSRecord(rec DNSRecord) CloudflareDNSRecord {
	return CloudflareDNSRecord{
		rec.dnsType,
		rec.url,
		rec.ip,
		120, // TODO Make these parameters configurable
		0,
		true,
	}
}

type CloudflareClient struct {
	baseURL string
	apiKey  string
	apiMail string
	client  http.Client
}

func NewCloudflareClient(baseURL, apiKey, apiMail string) CloudflareClient {
	return CloudflareClient{
		baseURL,
		apiKey,
		apiMail,
		http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
