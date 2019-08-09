package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	API_URL      string = "https://api.cloudflare.com/client/v4/"
	DNS_URL      string = "%szones/%s/dns_records"
	TOKEN_HEADER string = "Authorization"
)

// CloudflareDNSRecord represents a DNS record API object on cloudflare
// This structure will be serialized and send to the cloudflare api
type CloudflareDNSRecord struct {
	DnsType  string `json:"type"`
	Name     string `json:"name"`
	Content  net.IP `json:"content"`
	Ttl      uint16 `json:"ttl"`
	Priority uint16 `json:"priority"`
	Proxied  bool   `json:"proxied"`
}

func newFromDNSRecord(rec DNSRecord) CloudflareDNSRecord {
	return CloudflareDNSRecord{
		rec.DnsType,
		rec.Url,
		rec.Ip,
		120, // TODO Make these parameters configurable
		0,
		true,
	}
}

type CloudflareClient struct {
	apiToken string
	zone     string
	client   http.Client
}

func NewCloudflareClient(apiToken string, zone string) (CloudflareClient, error) {
	return CloudflareClient{
		apiToken,
		zone,
		http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (c *CloudflareClient) AddDNSRecord(record DNSRecord) error {
	cRecord := newFromDNSRecord(record)

	json, err := json.Marshal(cRecord)
	log.Println(string(json))
	if err != nil {
		return errors.Wrap(err, "Could not serialize json")
	}
	dnsUrl := fmt.Sprintf(DNS_URL, API_URL, c.zone)
	log.Println(dnsUrl)
	req, err := http.NewRequest("POST", dnsUrl, bytes.NewBuffer(json))
	if err != nil {
		return errors.Wrap(err, "Failed to build POST request to add DNS Record")
	}

	req.Header.Set(TOKEN_HEADER, fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed to make POST request to add DNS Record")
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(body))

	return nil
}
