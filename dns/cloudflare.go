package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	API_URL            string = "https://api.cloudflare.com/client/v4/"
	DNS_URL            string = "%szones/%s/dns_records"
	DNS_ID_URL         string = "%szones/%s/dns_records/%s"
	TOKEN_HEADER       string = "Authorization"
	ERR_ALREADY_EXISTS int    = 81057
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

func (c *CloudflareClient) ListDNSRecords() (*CloudflareDNSListResponse, error) {

	dnsUrl := fmt.Sprintf(DNS_URL, API_URL, c.zone)

	resp, err := c.doRequest("GET", dnsUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Request to list DNS records failed")
	}

	cfResp, err := parseIntoCloudflareResponseList(resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert response to Cloudflare response")
	}

	return cfResp, nil
}

func (c *CloudflareClient) AddDNSRecord(record DNSRecord) (*CloudflareResponse, error) {
	cRecord := newFromDNSRecord(record)

	dnsUrl := fmt.Sprintf(DNS_URL, API_URL, c.zone)

	resp, err := c.doRequest("POST", dnsUrl, cRecord)
	if err != nil {
		return nil, errors.Wrap(err, "Request to add DNS record failed")
	}

	cfResp, err := parseIntoCloudflareResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert response to Cloudflare response")
	}

	return cfResp, nil
}

func (c *CloudflareClient) DeleteDNSRecord(id string) (*CloudflareResponse, error) {

	dnsUrl := fmt.Sprintf(DNS_ID_URL, API_URL, c.zone, id)

	resp, err := c.doRequest("DELETE", dnsUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request to add DNS record failed")
	}

	cfResp, err := parseIntoCloudflareResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert response to Cloudflare response")
	}

	return cfResp, nil
}

func (c *CloudflareClient) doRequest(method, url string, payload interface{}) (*http.Response, error) {
	reqJson, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "Could not serialize json")
	}
	var (
		req *http.Request
	)
	if method != "GET" && method != "DELETE" {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqJson))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to build request")
	}

	req.Header.Set(TOKEN_HEADER, fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to make request")
	}
	return resp, nil
}

func parseIntoCloudflareResponse(resp *http.Response) (*CloudflareResponse, error) {
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response body")
	}
	var cfResponse CloudflareResponse
	err = json.Unmarshal(bytes, &cfResponse)
	if err != nil {
		return nil, errors.Wrap(err, "unable to deserialize response")
	}
	return &cfResponse, nil
}

func parseIntoCloudflareResponseList(resp *http.Response) (*CloudflareDNSListResponse, error) {
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response body")
	}
	var cfResponse CloudflareDNSListResponse
	err = json.Unmarshal(bytes, &cfResponse)
	if err != nil {
		return nil, errors.Wrap(err, "unable to deserialize response")
	}
	return &cfResponse, nil
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
