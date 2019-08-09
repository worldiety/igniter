package ingress

import (
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"
)

type IngressHolder struct {
	Id        types.UID
	Namespace string
	Name      string
	URLs      []string
}

func ParseIngressRespone(ingressList *v1beta1.IngressList) ([]IngressHolder, error) {
	var ret []IngressHolder
	for _, ingress := range ingressList.Items {
		ret = append(ret, ParseSingleIngress(&ingress))
	}
	return ret, nil
}

func ParseSingleIngress(ingress *v1beta1.Ingress) IngressHolder {
	id := ingress.UID
	namespace := ingress.Namespace
	name := ingress.Name

	return IngressHolder{id, namespace, name, mapUrls(ingress)}
}

func mapUrls(ingress *v1beta1.Ingress) []string {
	var urls []string
	for _, rule := range ingress.Spec.Rules {
		urls = append(urls, rule.Host)
	}
	return urls
}
