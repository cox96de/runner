package kube

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func ComposeKubeClientFromFile(kubeconfig string) (*kubernetes.Clientset, error) {
	// Creates the kubeconfig from the config file.
	// kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
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
