package main

type Config struct {
	Engine string
	Kube   struct {
		Config         string
		PortForwarding bool
		ExecutorImage  string
		ExecutorPath   string
		Namespace      string
	}
}
