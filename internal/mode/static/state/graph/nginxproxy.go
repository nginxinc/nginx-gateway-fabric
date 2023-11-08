package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

type NginxProxy struct {
	Source  *ngfAPI.NginxProxy
	Tracing *Tracing
}

// Tracing represents Otel tracing configuration for the dataplane.
type Tracing struct {
	// Endpoint specifies the address of OTLP/gRPC endpoint that will accept telemetry data.
	Endpoint string
	// ServiceName is the “service.name” attribute of the OTel resource.
	ServiceName string
	// Interval specifies the tracing interval.
	Interval string
	// BatchSize specifies the maximum number of spans to be sent in one batch per worker.
	BatchSize int
	// BatchCount specifies the number of pending batches per worker, spans exceeding the limit are dropped.
	BatchCount int
	// Enabled specifies if tracing is enabled.
	Enabled bool
}

func getNginxProxyConfig(
	nps map[types.NamespacedName]*ngfAPI.NginxProxy,
	gc *v1beta1.GatewayClass,
) *NginxProxy {
	if gc != nil {
		ref := gc.Spec.ParametersRef
		if ref != nil && ref.Namespace != nil &&
			ref.Group == ngfAPI.GroupName && ref.Kind == v1beta1.Kind("NginxProxy") {
			nsName := types.NamespacedName{Name: ref.Name, Namespace: string(*ref.Namespace)}
			return convertProxyConfig(nps[nsName], gc.Name)
		}
	}

	return nil
}

func convertProxyConfig(ngfproxy *ngfAPI.NginxProxy, gwcName string) *NginxProxy {
	if ngfproxy != nil {
		proxy := &NginxProxy{Source: ngfproxy}
		if ngfproxy.Spec.HTTP != nil && ngfproxy.Spec.HTTP.Telemetry != nil &&
			ngfproxy.Spec.HTTP.Telemetry.Tracing != nil {
			tracing := ngfproxy.Spec.HTTP.Telemetry.Tracing
			proxy.Tracing = &Tracing{
				Endpoint:    tracing.Endpoint,
				ServiceName: gwcName + ":ngf",
				Enabled:     tracing.Enabled,
				Interval:    tracing.Interval,
				BatchSize:   tracing.BatchSize,
				BatchCount:  tracing.BatchCount,
			}
		}
		return proxy
	}
	return nil
}
