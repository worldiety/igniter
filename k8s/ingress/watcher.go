package ingress

import (
	"fmt"
	"gitlab.worldiety.net/flahde/igniter/dns"
	"gitlab.worldiety.net/flahde/igniter/k8s/node"
	"k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
	"log"
	"sync"
	"time"
)

var (
	nodeInfos  []node.NodeInfo
	nodesMutex sync.RWMutex
	cfClient   dns.CloudflareClient
)

func WatchIngresses(clientset *kubernetes.Clientset, nodes []node.NodeInfo, cloudflareClient dns.CloudflareClient, done <-chan struct{}) {
	nodesMutex.Lock()
	nodeInfos = nodes
	nodesMutex.Unlock()
	cfClient = cloudflareClient

	watchList := cache.NewListWatchFromClient(clientset.ExtensionsV1beta1().RESTClient(), "ingresses", v1.NamespaceAll, fields.Everything())

	_, controller := cache.NewInformer(
		watchList,
		&v1beta1.Ingress{},
		30*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleIngressAdd,
			UpdateFunc: handleIngressUpdate,
			DeleteFunc: handleIngressDelete,
		},
	)
	go controller.Run(done)
}

func handleIngressAdd(obj interface{}) {
	nodesMutex.RLock()
	nodes := nodeInfos
	nodesMutex.RUnlock()

	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", obj)
		return
	}
	parsed := ParseSingleIngress(ingress)
	dnsRecords := dns.NewDNSRecords("A", parsed.URLs, nodes)
	for _, rec := range dnsRecords {
		resp, err := cfClient.AddDNSRecord(rec)
		if err != nil {
			log.Println("ERR: Error during request to cloudflare", err)
		} else if dns.IsAlreadyExistsError(resp) {
			log.Printf("Skipping DNS Record %s: Already exists", rec.Url)
		} else if dns.IsSuccess(resp) {
			log.Printf("Added DNS record for %s", rec.Url)
		} else {
			log.Println("got unknown error while POST request", resp)
		}
	}
}

func handleIngressUpdate(old, new interface{}) {
	nodesMutex.RLock()
	nodes := nodeInfos
	nodesMutex.RUnlock()

	oldIngress, ok := old.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", old)
		return
	}
	newIngress, ok := new.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", new)
		return
	}

	oldParsed := ParseSingleIngress(oldIngress)
	newParsed := ParseSingleIngress(newIngress)

	changed := false
	for i, oldUrl := range oldParsed.URLs {
		if oldUrl != newParsed.URLs[i] {
			changed = true
			break
		}
	}

	if !changed {
		log.Printf("Skipping updating DNS Record %s: Hasn't changed", oldParsed.URLs[0])
		return
	}

	oldDNSRecords := dns.NewDNSRecords("A", oldParsed.URLs, nodes)
	newDNSRecords := dns.NewDNSRecords("A", newParsed.URLs, nodes)
	listResp, err := cfClient.ListDNSRecords()
	if err != nil {
		log.Println("ERR: unable to get DNS records from Cloudflare", err)
		return
	}
	for i, rec := range newDNSRecords {
		id, err := getIdForDNSRecord(oldDNSRecords[i], listResp)
		if err != nil {
			log.Println(err)
			continue
		}
		resp, err := cfClient.UpdateDNSRecord(id, rec)
		if err != nil {
			log.Println("ERR: Error during request to cloudflare", err)
		} else if dns.IsAlreadyExistsError(resp) {
			log.Printf("Skipping DNS Record %s: Already exists", rec.Url)
		} else if dns.IsSuccess(resp) {
			log.Printf("Updated DNS record for %s", rec.Url)
		} else {
			log.Println("Got unknown error while POST request", resp)
		}
	}
}

func handleIngressDelete(obj interface{}) {
	nodesMutex.RLock()
	nodes := nodeInfos
	nodesMutex.RUnlock()

	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", obj)
		return
	}
	parsed := ParseSingleIngress(ingress)

	dnsRecords := dns.NewDNSRecords("A", parsed.URLs, nodes)
	listResp, err := cfClient.ListDNSRecords()
	if err != nil {
		log.Println("ERR: unable to get DNS records from Cloudflare", err)
	}
	for _, rec := range dnsRecords {
		id, err := getIdForDNSRecord(rec, listResp)
		if err != nil {
			log.Println(err)
			continue
		}

		resp, err := cfClient.DeleteDNSRecord(id)
		if err != nil {
			log.Println("ERR: Error during request to cloudflare", err)
		} else if dns.IsSuccess(resp) {
			log.Printf("Deleted DNS Record for %s", rec.Url)
		} else {
			log.Println("got unknown error while DELETE request", resp)
		}
	}
}

func getIdForDNSRecord(record dns.DNSRecord, response *dns.CloudflareDNSListResponse) (string, error) {
	for _, cfRecord := range response.Result {
		if cfRecord.Name == record.Url && cfRecord.Content == record.Ip.String() {
			return cfRecord.ID, nil
		}
	}
	return "", fmt.Errorf("couldn't find matching cloudflare dns record for %s", record.Url)
}
