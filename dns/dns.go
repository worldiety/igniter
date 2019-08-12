package dns

import (
	"fmt"
	"gitlab.worldiety.net/flahde/igniter/k8s/node"
	"net"
)

// DNSRecord is a representation of a DNS record
type DNSRecord struct {
	DnsType string
	Url     string
	Ip      net.IP
}

func (record DNSRecord) String() string {
	return fmt.Sprintf("%s %s %s", record.DnsType, record.Url, record.Ip.String())
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
