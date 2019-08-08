package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.worldiety.net/flahde/igniter/k8s/ingress"
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

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	interrupt := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(interrupt, os.Interrupt)

	fmt.Println("Starting watcher")
	ingress.WatchIngresses(clientset, done)
	for sig := range interrupt {
		fmt.Printf("Recieved %v, stopping\n", sig)
		var s struct{}
		done <- s
		break
	}
}
