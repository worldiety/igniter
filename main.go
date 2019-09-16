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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/signal"
	"syscall"
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

func cloudflareZone() (string, error) {
	if token := os.Getenv("CLOUDFLARE_ZONE"); token != "" {
		return token, nil
	}
	return "", fmt.Errorf("Could not find cloudflare zone identifier in environment")
}

func main() {
	var (
		kubeconfig     *string
		outOfCluster   *bool
		shouldProxyDNS *bool
	)
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	outOfCluster = flag.Bool("outofcluster", false, "(optional) set this to true if testing on a dev machine")
	shouldProxyDNS = flag.Bool("proxy", true, "(optional) controls wether records will be proxied through Clouflare or not. Currently existing records are not affected")
	flag.Parse()

	cloudflareToken, err := cloudflareToken()
	if err != nil {
		log.Fatal(err)
	}

	cloudflareZone, err := cloudflareZone()
	if err != nil {
		log.Fatal(err)
	}

	var config *rest.Config
	if *outOfCluster {
		log.Printf("Running in out-cluster-mode using %s", *kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		log.Println("Running in in-cluster mode")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodes, err := node.GetInfoAboutWorkerNodes(clientset)
	if err != nil {
		log.Fatal(err)
	}

	for _, n := range nodes {
		log.Printf("Found node '%s' with IP '%s'", n.Name, n.PublicIP)
	}

	log.Printf("Building Cloudflare client with proxy = %t", *shouldProxyDNS)
	cloudflareClient, err := cloudflare.NewCloudflareClient(cloudflareToken, cloudflareZone, *shouldProxyDNS)
	if err != nil {
		log.Fatal("Failed to build Cloudflare client", err)
	}

	log.Println("Starting watcher")
	done := make(chan struct{}, 1)
	ingress.WatchIngresses(clientset, nodes, cloudflareClient, done)

	interrupt := make(chan os.Signal, 1)
	if *outOfCluster {
		log.Println("Press Ctrl-c to stop")
		signal.Notify(interrupt, os.Interrupt)
	} else {
		// k8s tries to stop pods gracefully by sending SIGTERM, so let's listen for that
		signal.Notify(interrupt, syscall.SIGTERM)
	}

	for sig := range interrupt {
		log.Printf("Recieved %v, stopping\n", sig)
		var s struct{}
		done <- s
		break
	}
}
