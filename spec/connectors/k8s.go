package connectors

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	api "github.com/vultr/slinkee/spec/api/types/v1"
	client "github.com/vultr/slinkee/spec/clientset/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// GetKubernetesConn returns a connection to the kubernetes cluster
func GetKubernetesConn() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// GetSlinkeeClientset returns a slinkee clientset to interact with the k8s cluster
func GetSlinkeeClientset() (*client.V1Client, error) {
	if err := api.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := client.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
