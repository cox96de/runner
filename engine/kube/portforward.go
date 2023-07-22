package kube

import (
	"io"
	"net/http"

	"github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func newPortForward(cli kubernetes.Interface, config *rest.Config, namespace string, pod string, ports []string,
	stopChan <-chan struct{}, out,
	errOut io.Writer,
) (*portforward.PortForwarder, error) {
	req := cli.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("portforward")
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, "POST", req.URL())
	ready := make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer, ports, stopChan, ready, out, errOut)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return forwarder, nil
}
