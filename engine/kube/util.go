package kube

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ComposeKubeClientFromFile creates a kubernetes client from a kubeconfig file.
func ComposeKubeClientFromFile(kubeconfig string) (*kubernetes.Clientset, error) {
	// Creates the kubeconfig from the config file.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return clientset, err
}

// ComposeKubeClientInKube creates a kubernetes client in a k8s pod.
func ComposeKubeClientInKube() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
