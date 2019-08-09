package ingress

import (
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
)

func WatchIngresses(clientset *kubernetes.Clientset, nodes []node.NodeInfo, done <-chan struct{}) {
	nodesMutex.Lock()
	nodeInfos = nodes
	nodesMutex.Unlock()

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

	log.Printf("ADD %+v\n", dnsRecords)
}

func handleIngressUpdate(old, new interface{}) {
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
	log.Printf("UPDATE %v %v\n", ParseSingleIngress(oldIngress), ParseSingleIngress(newIngress))
}

func handleIngressDelete(obj interface{}) {
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", obj)
		return
	}
	log.Printf("DELETE %v\n", ParseSingleIngress(ingress))
}
