package k8s

import (
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

type IngressHolder struct {
	namespace string
	name      string
	urls      []string
}

func ParseIngressRespone(ingressList *v1beta1.IngressList) ([]IngressHolder, error) {
	var ret []IngressHolder
	return ret, nil
}
