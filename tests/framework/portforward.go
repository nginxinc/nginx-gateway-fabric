package framework

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNGFPodName returns the name of the NGF Pod.
func GetNGFPodName(
	k8sClient client.Client,
	namespace,
	releaseName string,
	timeout time.Duration,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var podList core.PodList
	if err := k8sClient.List(
		ctx,
		&podList,
		client.InNamespace(namespace),
		client.MatchingLabels{
			"app.kubernetes.io/instance": releaseName,
		},
	); err != nil {
		return "", fmt.Errorf("error getting list of Pods: %w", err)
	}

	if len(podList.Items) > 0 {
		return podList.Items[0].Name, nil
	}

	return "", fmt.Errorf("unable to find NGF Pod")
}

// PortForward starts a port forward to the specified Pod and returns the local port being forwarded.
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
