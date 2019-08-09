package node

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
	"strings"
)

// NodeInfo contains all informations about a node that we need
// to populate Cloudflare DNS Entries
type NodeInfo struct {
	Name     string
	PublicIP net.IP
}

func (node NodeInfo) String() string {
	return node.Name + "  " + node.PublicIP.String()
}

func GetInfoAboutWorkerNodes(clientset *kubernetes.Clientset) ([]NodeInfo, error) {
	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "could not get information about nodes")
	}
	ret := make([]NodeInfo, 0)
	for _, node := range nodes.Items {
		ip := node.Annotations["flannel.alpha.coreos.com/public-ip"]
		name := node.Name

		if strings.Contains(name, "master") {
			continue
		}

		info := NodeInfo{
			name,
			net.ParseIP(ip),
		}
		ret = append(ret, info)
	}

	return ret, nil
}
