package dns

import (
	"fmt"
	"gitlab.worldiety.net/flahde/igniter/k8s/node"
	"net"
)

// DNSRecord is a representation of a DNS record
type DNSRecord struct {
	dnsType string
	url     string
	ip      net.IP
}

func (record DNSRecord) String() string {
	return fmt.Sprintf("%s %s %s", record.dnsType, record.url, record.ip.String())
}

func NewDNSRecords(dnsType string, urls []string, nodes []node.NodeInfo) []DNSRecord {
	ret := make([]DNSRecord, 0)
	for _, url := range urls {
		for _, node := range nodes {
			record := DNSRecord{
				dnsType,
				url,
				node.PublicIP,
			}
			ret = append(ret, record)
		}
	}

	return ret
}

type DNSHandler interface {
	// HasDNSRecord checks wether the given DNS record exists on cloudflare
	HasDNSRecord(DNSRecord) (bool, error)
	// AddDNSRecord adds an DNS record to cloudflare. TODO: What happens when called twice with the same record?
	AddDNSRecord(DNSRecord) error
	// RemoveDNSRecord deletes an DNS record on cloudflare. TODO: What happens when called twice with the same record?
	RemoveDNSRecord(DNSRecord) error
}
