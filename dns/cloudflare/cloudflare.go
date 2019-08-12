package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.worldiety.net/flahde/igniter/dns"
	"io/ioutil"
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

func (c *CloudflareClient) ListDNSRecords() (*CloudflareDNSListResponse, error) {

	dnsUrl := fmt.Sprintf(DNS_URL, API_URL, c.zone) + "?per_page=100"

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

func (c *CloudflareClient) AddDNSRecord(record dns.DNSRecord) (*CloudflareResponse, error) {
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

func (c *CloudflareClient) UpdateDNSRecord(id string, record dns.DNSRecord) (*CloudflareResponse, error) {
	cRecord := newFromDNSRecord(record)
	dnsUrl := fmt.Sprintf(DNS_ID_URL, API_URL, c.zone, id)

	resp, err := c.doRequest("PUT", dnsUrl, cRecord)
	if err != nil {
		return nil, errors.Wrap(err, "request to delete record failed")
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
		return nil, errors.Wrap(err, "request to delete record failed")
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
