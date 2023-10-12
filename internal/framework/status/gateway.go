package status

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
)

// GetGatewayAddresses gets the addresses for the Gateway.
func GetGatewayAddresses(
	ctx context.Context,
	k8sClient client.Client,
	svc *v1.Service,
	podConfig config.GatewayPodConfig,
) ([]v1beta1.GatewayStatusAddress, error) {
	podAddress := []v1beta1.GatewayStatusAddress{
		{
			Type:  helpers.GetPointer(v1beta1.IPAddressType),
			Value: podConfig.PodIP,
		},
	}

	var gwSvc v1.Service
	if svc == nil {
		key := types.NamespacedName{Name: podConfig.ServiceName, Namespace: podConfig.Namespace}
		if err := k8sClient.Get(ctx, key, &gwSvc); err != nil {
			return podAddress, fmt.Errorf("error finding Service for Gateway: %w", err)
		}
	} else {
		gwSvc = *svc
	}

	var addresses, hostnames []string
	switch gwSvc.Spec.Type {
	case v1.ServiceTypeNodePort:
		var err error
		addresses, err = getNodeAddresses(ctx, k8sClient)
		if err != nil {
			return podAddress, fmt.Errorf("error getting Node addresses: %w", err)
		}
	case v1.ServiceTypeLoadBalancer:
		for _, ingress := range gwSvc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				addresses = append(addresses, ingress.IP)
			} else if ingress.Hostname != "" {
				hostnames = append(hostnames, ingress.Hostname)
			}
		}
	}

	gwAddresses := make([]v1beta1.GatewayStatusAddress, 0, len(addresses)+len(hostnames))
	for _, addr := range addresses {
		statusAddr := v1beta1.GatewayStatusAddress{
			Type:  helpers.GetPointer(v1beta1.IPAddressType),
			Value: addr,
		}
		gwAddresses = append(gwAddresses, statusAddr)
	}

	for _, hostname := range hostnames {
		statusAddr := v1beta1.GatewayStatusAddress{
			Type:  helpers.GetPointer(v1beta1.HostnameAddressType),
			Value: hostname,
		}
		gwAddresses = append(gwAddresses, statusAddr)
	}

	return gwAddresses, nil
}

func getNodeAddresses(
	ctx context.Context,
	k8sClient client.Client,
) ([]string, error) {
	var nodeList v1.NodeList
	if err := k8sClient.List(ctx, &nodeList); err != nil {
		return nil, err
	}

	nodeAddresses := make([]string, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		var externalIP, internalIP string
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeExternalIP {
				externalIP = address.Address
			}
			if address.Type == v1.NodeInternalIP {
				internalIP = address.Address
			}
		}
		if externalIP != "" {
			nodeAddresses = append(nodeAddresses, externalIP)
		} else if internalIP != "" {
			nodeAddresses = append(nodeAddresses, internalIP)
		}
	}

	return nodeAddresses, nil
}

// prepareGatewayStatus prepares the status for a Gateway resource.
func prepareGatewayStatus(
	gatewayStatus GatewayStatus,
	transitionTime metav1.Time,
) v1beta1.GatewayStatus {
	listenerStatuses := make([]v1beta1.ListenerStatus, 0, len(gatewayStatus.ListenerStatuses))

	// FIXME(pleshakov) Maintain the order from the Gateway resource
	// https://github.com/nginxinc/nginx-gateway-fabric/issues/689
	names := make([]string, 0, len(gatewayStatus.ListenerStatuses))
	for name := range gatewayStatus.ListenerStatuses {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		s := gatewayStatus.ListenerStatuses[name]

		listenerStatuses = append(listenerStatuses, v1beta1.ListenerStatus{
			Name:           v1beta1.SectionName(name),
			SupportedKinds: s.SupportedKinds,
			AttachedRoutes: s.AttachedRoutes,
			Conditions:     convertConditions(s.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
		})
	}

	return v1beta1.GatewayStatus{
		Listeners:  listenerStatuses,
		Addresses:  gatewayStatus.Addresses,
		Conditions: convertConditions(gatewayStatus.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
	}
}
