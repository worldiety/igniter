package cloudflare

import (
	"gitlab.worldiety.net/flahde/igniter/dns"
	"net"
	"time"
)

const (
	DEFAULT_TTL              = 120
	DEFAULT_CLOUDFLARE_PROXY = true
)

// CloudflareDNSRecord represents a DNS record API object on cloudflare
// This structure will be serialized and send to the cloudflare api
type CloudflareDNSRecord struct {
	DnsType string `json:"type"`
	Name    string `json:"name"`
	Content net.IP `json:"content"`
	Ttl     uint16 `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func newFromDNSRecord(rec dns.DNSRecord) CloudflareDNSRecord {
	return CloudflareDNSRecord{
		rec.DnsType,
		rec.Url,
		rec.Ip,
		DEFAULT_TTL,
		DEFAULT_CLOUDFLARE_PROXY,
	}
}

type CloudflareResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   interface{}   `json:"result"`
}

type CloudflareDNSListResponse struct {
	Result []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		ModifiedOn time.Time `json:"modified_on"`
		CreatedOn  time.Time `json:"created_on"`
		Meta       struct {
			AutoAdded           bool `json:"auto_added"`
			ManagedByApps       bool `json:"managed_by_apps"`
			ManagedByArgoTunnel bool `json:"managed_by_argo_tunnel"`
		} `json:"meta"`
	} `json:"result"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

func IsAlreadyExistsError(resp *CloudflareResponse) bool {
	if len(resp.Errors) == 0 {
		return false

	}
	return resp.Errors[0].Code == ERR_ALREADY_EXISTS
}

func IsSuccess(resp *CloudflareResponse) bool {
	return len(resp.Errors) == 0
}
