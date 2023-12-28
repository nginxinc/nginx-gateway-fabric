// Utility functions for managing resources in Kubernetes. Inspiration and methods used from
// https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/utils.

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// ResourceManager handles creating/updating/deleting Kubernetes resources.
type ResourceManager struct {
	K8sClient     client.Client
	FS            embed.FS
	TimeoutConfig TimeoutConfig
}

// ClusterInfo holds the cluster metadata
type ClusterInfo struct {
	K8sVersion      string
	MemoryPerNode   string
	GkeInstanceType string
	GkeZone         string
	NodeCount       int
	CPUCountPerNode int64
	MaxPodsPerNode  int64
	IsGKE           bool
}

// Apply creates or updates Kubernetes resources defined as Go objects.
func (rm *ResourceManager) Apply(resources []client.Object) error {
	ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.CreateTimeout)
	defer cancel()

	for _, resource := range resources {
		if err := rm.K8sClient.Get(ctx, client.ObjectKeyFromObject(resource), resource); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("error getting resource: %w", err)
			}

			if err := rm.K8sClient.Create(ctx, resource); err != nil {
				return fmt.Errorf("error creating resource: %w", err)
			}

			continue
		}

		if err := rm.K8sClient.Update(ctx, resource); err != nil {
			return fmt.Errorf("error updating resource: %w", err)
		}
	}

	return nil
}

// ApplyFromFiles creates or updates Kubernetes resources defined within the provided YAML files.
func (rm *ResourceManager) ApplyFromFiles(files []string, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.CreateTimeout)
	defer cancel()

	handlerFunc := func(obj unstructured.Unstructured) error {
		obj.SetNamespace(namespace)
		nsName := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
		fetchedObj := obj.DeepCopy()
		if err := rm.K8sClient.Get(ctx, nsName, fetchedObj); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("error getting resource: %w", err)
			}

			if err := rm.K8sClient.Create(ctx, &obj); err != nil {
				return fmt.Errorf("error creating resource: %w", err)
			}

			return nil
		}

		obj.SetResourceVersion(fetchedObj.GetResourceVersion())
		if err := rm.K8sClient.Update(ctx, &obj); err != nil {
			return fmt.Errorf("error updating resource: %w", err)
		}

		return nil
	}

	return rm.readAndHandleObjects(handlerFunc, files)
}

// Delete deletes Kubernetes resources defined as Go objects.
func (rm *ResourceManager) Delete(resources []client.Object) error {
	for _, resource := range resources {
		ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.DeleteTimeout)
		defer cancel()

		if err := rm.K8sClient.Delete(ctx, resource); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("error deleting resource: %w", err)
		}
	}

	return nil
}

// DeleteFromFile deletes Kubernetes resources defined within the provided YAML files.
func (rm *ResourceManager) DeleteFromFiles(files []string, namespace string) error {
	handlerFunc := func(obj unstructured.Unstructured) error {
		obj.SetNamespace(namespace)
		ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.DeleteTimeout)
		defer cancel()

		if err := rm.K8sClient.Delete(ctx, &obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		return nil
	}

	return rm.readAndHandleObjects(handlerFunc, files)
}

func (rm *ResourceManager) readAndHandleObjects(
	handle func(unstructured.Unstructured) error,
	files []string,
) error {
	for _, file := range files {
		data, err := rm.getFileContents(file)
		if err != nil {
			return err
		}

		decoder := yaml.NewYAMLOrJSONDecoder(data, 4096)
		for {
			obj := unstructured.Unstructured{}
			if err := decoder.Decode(&obj); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return fmt.Errorf("error decoding resource: %w", err)
			}

			if len(obj.Object) == 0 {
				continue
			}

			if err := handle(obj); err != nil {
				return err
			}
		}
	}

	return nil
}

// getFileContents takes a string that can either be a local file
// path or an https:// URL to YAML manifests and provides the contents.
func (rm *ResourceManager) getFileContents(file string) (*bytes.Buffer, error) {
	if strings.HasPrefix(file, "http://") {
		return nil, fmt.Errorf("data can't be retrieved from %s: http is not supported, use https", file)
	} else if strings.HasPrefix(file, "https://") {
		ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.ManifestFetchTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, file, nil)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%d response when getting %s file contents", resp.StatusCode, file)
		}

		manifests := new(bytes.Buffer)
		count, err := manifests.ReadFrom(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.ContentLength != -1 && count != resp.ContentLength {
			return nil, fmt.Errorf("received %d bytes from %s, expected %d", count, file, resp.ContentLength)
		}
		return manifests, nil
	}

	if !strings.HasPrefix(file, "manifests/") {
		file = "manifests/" + file
	}

	b, err := rm.FS.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

// WaitForAppsToBeReady waits for all apps in the specified namespace to be ready,
// or until the ctx timeout is reached.
func (rm *ResourceManager) WaitForAppsToBeReady(namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.CreateTimeout)
	defer cancel()

	if err := rm.waitForPodsToBeReady(ctx, namespace); err != nil {
		return err
	}

	if err := rm.waitForRoutesToBeReady(ctx, namespace); err != nil {
		return err
	}

	return rm.waitForGatewaysToBeReady(ctx, namespace)
}

func (rm *ResourceManager) waitForPodsToBeReady(ctx context.Context, namespace string) error {
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var podList core.PodList
			if err := rm.K8sClient.List(ctx, &podList, client.InNamespace(namespace)); err != nil {
				return false, err
			}

			var podsReady int
			for _, pod := range podList.Items {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == core.PodReady && cond.Status == core.ConditionTrue {
						podsReady++
					}
				}
			}

			return podsReady == len(podList.Items), nil
		},
	)
}

func (rm *ResourceManager) waitForGatewaysToBeReady(ctx context.Context, namespace string) error {
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var gatewayList v1.GatewayList
			if err := rm.K8sClient.List(ctx, &gatewayList, client.InNamespace(namespace)); err != nil {
				return false, err
			}

			for _, gw := range gatewayList.Items {
				for _, cond := range gw.Status.Conditions {
					if cond.Type == string(v1.GatewayConditionProgrammed) && cond.Status == metav1.ConditionTrue {
						return true, nil
					}
				}
			}

			return false, nil
		},
	)
}

func (rm *ResourceManager) waitForRoutesToBeReady(ctx context.Context, namespace string) error {
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var routeList v1.HTTPRouteList
			if err := rm.K8sClient.List(ctx, &routeList, client.InNamespace(namespace)); err != nil {
				return false, err
			}

			var numParents, readyCount int
			for _, route := range routeList.Items {
				numParents += len(route.Status.Parents)
				for _, parent := range route.Status.Parents {
					for _, cond := range parent.Conditions {
						if cond.Type == string(v1.RouteConditionAccepted) && cond.Status == metav1.ConditionTrue {
							readyCount++
						}
					}
				}
			}

			return numParents == readyCount, nil
		},
	)
}

// GetLBIPAddress gets the IP or Hostname from the Loadbalancer service.
func (rm *ResourceManager) GetLBIPAddress(namespace string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.CreateTimeout)
	defer cancel()

	var serviceList core.ServiceList
	var address string
	if err := rm.K8sClient.List(ctx, &serviceList, client.InNamespace(namespace)); err != nil {
		return "", err
	}
	var nsName types.NamespacedName

	for _, svc := range serviceList.Items {
		if svc.Spec.Type == core.ServiceTypeLoadBalancer {
			nsName = types.NamespacedName{Namespace: svc.GetNamespace(), Name: svc.GetName()}
			if err := rm.waitForLBStatusToBeReady(ctx, nsName); err != nil {
				return "", fmt.Errorf("error getting status from LoadBalancer service: %w", err)
			}
		}
	}

	if nsName.Name != "" {
		var lbService core.Service

		if err := rm.K8sClient.Get(ctx, nsName, &lbService); err != nil {
			return "", fmt.Errorf("error getting LoadBalancer service: %w", err)
		}
		if lbService.Status.LoadBalancer.Ingress[0].IP != "" {
			address = lbService.Status.LoadBalancer.Ingress[0].IP
		} else if lbService.Status.LoadBalancer.Ingress[0].Hostname != "" {
			address = lbService.Status.LoadBalancer.Ingress[0].Hostname
		}
		return address, nil
	}
	return "", nil
}

func (rm *ResourceManager) waitForLBStatusToBeReady(ctx context.Context, svcNsName types.NamespacedName) error {
	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var svc core.Service
			if err := rm.K8sClient.Get(ctx, svcNsName, &svc); err != nil {
				return false, err
			}
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				return true, nil
			}

			return false, nil
		},
	)
}

// GetClusterInfo retrieves node info and Kubernetes version from the cluster
func (rm *ResourceManager) GetClusterInfo() (ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rm.TimeoutConfig.GetTimeout)
	defer cancel()

	var nodes core.NodeList
	ci := &ClusterInfo{}
	if err := rm.K8sClient.List(ctx, &nodes); err != nil {
		return *ci, fmt.Errorf("error getting nodes: %w", err)
	}

	ci.NodeCount = len(nodes.Items)

	node := nodes.Items[0]
	ci.K8sVersion = node.Status.NodeInfo.KubeletVersion
	ci.CPUCountPerNode, _ = node.Status.Capacity.Cpu().AsInt64()
	ci.MemoryPerNode = node.Status.Capacity.Memory().String()
	ci.MaxPodsPerNode, _ = node.Status.Capacity.Pods().AsInt64()
	providerID := node.Spec.ProviderID

	if strings.Split(providerID, "://")[0] == "gce" {
		ci.IsGKE = true
		ci.GkeInstanceType = node.Labels["beta.kubernetes.io/instance-type"]
		ci.GkeZone = node.Labels["topology.kubernetes.io/zone"]
	}

	return *ci, nil
}

// GetNGFPodNames returns the name(s) of the NGF Pod(s).
func GetNGFPodNames(
	k8sClient client.Client,
	namespace,
	releaseName string,
	timeout time.Duration,
) ([]string, error) {
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
		return nil, fmt.Errorf("error getting list of Pods: %w", err)
	}

	if len(podList.Items) > 0 {
		var names []string
		for _, pod := range podList.Items {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == core.PodReady && cond.Status == core.ConditionTrue {
					names = append(names, pod.Name)
				}
			}
		}
		return names, nil
	}

	return nil, errors.New("unable to find NGF Pod(s)")
}
