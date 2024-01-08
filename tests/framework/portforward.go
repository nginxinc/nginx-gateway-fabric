package framework

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForward starts a port-forward to the specified Pod and returns the local port being forwarded.
func PortForward(config *rest.Config, namespace, podName string, stopCh chan struct{}) (int, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return 0, fmt.Errorf("error creating roundtripper: %w", err)
	}

	serverURL, err := url.Parse(config.Host)
	if err != nil {
		return 0, fmt.Errorf("error parsing rest config host: %w", err)
	}

	serverURL.Path = path.Join(
		"api", "v1",
		"namespaces", namespace,
		"pods", podName,
		"portforward",
	)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	readyCh := make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{":80"}, stopCh, readyCh, out, errOut)
	if err != nil {
		return 0, fmt.Errorf("error creating port forwarder: %w", err)
	}

	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			panic(err)
		}
	}()

	<-readyCh
	ports, err := forwarder.GetPorts()
	if err != nil {
		return 0, fmt.Errorf("error getting ports being forwarded: %w", err)
	}

	return int(ports[0].Local), nil
}
