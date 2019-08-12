package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gitlab.worldiety.net/flahde/igniter/dns/cloudflare"
	"gitlab.worldiety.net/flahde/igniter/k8s/ingress"
	"gitlab.worldiety.net/flahde/igniter/k8s/node"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os/signal"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func cloudflareToken() (string, error) {
	if token := os.Getenv("CLOUDFLARE_API_TOKEN"); token != "" {
		return token, nil
	}
	return "", fmt.Errorf("Could not find cloudflare api token in environment")
}

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	cloudflareToken, err := cloudflareToken()
	if err != nil {
		log.Fatal(err)
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodes, err := node.GetInfoAboutWorkerNodes(clientset)
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		log.Printf("Found node '%s' with IP '%s'", node.Name, node.PublicIP)
	}

	cloudflareClient, err := cloudflare.NewCloudflareClient(cloudflareToken, "81664077c0050a0f3a8996c0402b8574")
	if err != nil {
		log.Fatal("Failed to build Cloudflare client", err)
	}
	//os.Exit(0)

	log.Println("Starting watcher")
	done := make(chan struct{}, 1)
	ingress.WatchIngresses(clientset, nodes, cloudflareClient, done)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for sig := range interrupt {
		log.Printf("Recieved %v, stopping\n", sig)
		var s struct{}
		done <- s
		break
	}
}
