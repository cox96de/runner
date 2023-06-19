//go:build kube_integration

package kube

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This example shows how to use the kube engine.
func ExampleEngine() {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}
	clientset, err := ComposeKubeClientFromFile(kubeconfig)
	checkError(err)
	namespace := "runner"
	// Create namespace, it should be created by the user.
	clientset.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	engine := NewEngine(clientset, &Option{
		ExecutorImage: "docker.io/cox96de/runner-executor:master",
		ExecutorPath:  "/executor",
		Namespace:     namespace,
	})
	err = engine.Ping(context.Background())
	checkError(err)
	// Output:
}

func checkError(err error) {
	if err == nil {
		return
	}
	panic(err)
}
