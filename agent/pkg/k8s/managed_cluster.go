package k8s

import (
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

func GetKubernetesVersion(client kubernetes.Interface) (*version.Info, error) {
	return client.Discovery().ServerVersion()
}
