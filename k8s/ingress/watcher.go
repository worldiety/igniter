package ingress

import (
	"fmt"
	"k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
	"log"
	"time"
)

func WatchIngresses(clientset *kubernetes.Clientset, done <-chan struct{}) {
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
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", obj)
		return
	}
	fmt.Printf("ADD %v\n", ParseSingleIngress(ingress))
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
	fmt.Printf("UPDATE %v %v\n", ParseSingleIngress(oldIngress), ParseSingleIngress(newIngress))
}

func handleIngressDelete(obj interface{}) {
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		log.Printf("Recieved unknown type: %t\n", obj)
		return
	}
	fmt.Printf("DELETE %v\n", ParseSingleIngress(ingress))
}
