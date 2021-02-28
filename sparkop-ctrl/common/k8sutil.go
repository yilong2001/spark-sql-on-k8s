package common

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func BuildK8sRestConfig(master, kubeConfig string) (*rest.Config, error) {
	if kubeConfig != "" {
		return clientcmd.BuildConfigFromFlags(master, kubeConfig)
	}
	return rest.InClusterConfig()
}

