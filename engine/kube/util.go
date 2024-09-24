package kube

import (
	"github.com/cockroachdb/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ComposeKubeClientFromFile creates a kubernetes client from a kubeconfig file.
func ComposeKubeClientFromFile(kubeconfig string) (*kubernetes.Clientset, *rest.Config, error) {
	// Creates the kubeconfig from the config file.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return clientset, config, err
}

// ComposeKubeClientInKube creates a kubernetes client in a k8s pod.
func ComposeKubeClientInKube() (*kubernetes.Clientset, *rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, config, err
}
