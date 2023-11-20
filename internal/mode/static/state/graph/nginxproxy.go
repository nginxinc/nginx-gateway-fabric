package graph

import (
	"regexp"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

// NginxProxy represents the user specified NGINX configuration features.
type NginxProxy struct {
	// Source is the corresponding NginxProxy proxy resource.
	Source *ngfAPI.NginxProxy
	// Tracing holds the tracing configuration.
	Tracing *Tracing
	// Conditions holds the conditions for the NginxProxy.
	Conditions []conditions.Condition
	// Valid indicates whether the NginxProxy Spec is valid.
	Valid bool
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
	BatchSize int32
	// BatchCount specifies the number of pending batches per worker, spans exceeding the limit are dropped.
	BatchCount int32
	// Enabled specifies if tracing is enabled.
	Enabled bool
}

func buildNginxProxyConfig(
	nps map[types.NamespacedName]*ngfAPI.NginxProxy,
	gc *v1.GatewayClass,
	gw *v1.Gateway,
) *NginxProxy {
	gwNsName := "unknown"
	if gw != nil {
		gwNsName = gw.Namespace + "/" + gw.Name
	}
	if gc != nil {
		ref := gc.Spec.ParametersRef
		if ref != nil && ref.Namespace != nil &&
			ref.Group == ngfAPI.GroupName && ref.Kind == v1.Kind("NginxProxy") {
			nsName := types.NamespacedName{Name: ref.Name, Namespace: string(*ref.Namespace)}
			return convertProxyConfig(nps[nsName], gwNsName)
		}
	}

	return nil
}

func convertProxyConfig(ngfproxy *ngfAPI.NginxProxy, gwNsName string) *NginxProxy {
	if ngfproxy != nil {
		return validateProxyConfig(ngfproxy, gwNsName)
	}
	return nil
}

func validateProxyConfig(ngfproxy *ngfAPI.NginxProxy, gwNsName string) *NginxProxy {
	var conds []conditions.Condition

	var valErr []error

	proxy := &NginxProxy{Source: ngfproxy}

	if ngfproxy.Spec.HTTP != nil && ngfproxy.Spec.HTTP.Telemetry != nil &&
		ngfproxy.Spec.HTTP.Telemetry.Tracing != nil {
		ngfTracing := ngfproxy.Spec.HTTP.Telemetry.Tracing
		tracing, valErr := validateTracing(ngfTracing, gwNsName)
		if len(valErr) == 0 {
			proxy.Tracing = tracing
		}
	}

	for _, ve := range valErr {
		conds = append(conds, staticConds.NewNginxProxyNotProgrammed(ve.Error()))
	}

	proxy.Conditions = conds
	proxy.Valid = len(valErr) == 0

	return proxy
}

const (
	endpointRegex            = `^(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}|(?:\d{1,3}\.){3}\d{1,3}):\d{1,5}$` // nolint:lll
	timeRegex                = `^(\d+y)??\s*(\d+M)??\s*(\d+w)??\s*(\d+d)??\s*(\d+h)??\s*(\d+m)??\s*(\d+s?)??\s*(\d+ms)??$`
	endpointValidationString = "tracing.endpoint must be in the form `<IP>:<port>` or `<hostname>:<port>"
)

var (
	reEndpoint = regexp.MustCompile(endpointRegex)
	reTime     = regexp.MustCompile(timeRegex)
)

func validateTracing(ngfTracing *ngfAPI.Tracing, gwNsName string) (*Tracing, []error) {
	var valErr []error
	var tracing *Tracing
	path := field.NewPath("spec.HTTP.telemetry").Child("tracing")
	if ngfTracing.Endpoint == "" {
		valErr = append(valErr, field.Required(path.Child("endpoint"), "tracing.endpoint must be specified"))
	} else if !reEndpoint.MatchString(ngfTracing.Endpoint) {
		valErr = append(valErr, field.Invalid(
			path.Child("endpoint"), ngfTracing.Endpoint, endpointValidationString))
	}
	if ngfTracing.Interval != nil && !reTime.MatchString(*ngfTracing.Interval) {
		valErr = append(valErr, field.Invalid(
			path.Child("interval"), *ngfTracing.Interval, "tracing.interval must be a valid time string"))
	}
	if ngfTracing.BatchCount != nil && *ngfTracing.BatchCount < 1 {
		valErr = append(valErr, field.Invalid(
			path.Child("batchCount"), *ngfTracing.BatchCount, "tracing.batchCount must be a positive integer"))
	}
	if ngfTracing.BatchSize != nil && *ngfTracing.BatchSize < 1 {
		valErr = append(valErr, field.Invalid(
			path.Child("batchSize"), *ngfTracing.BatchSize, "tracing.batchSize must be a positive integer"))
	}

	if len(valErr) == 0 {
		tracing = &Tracing{
			Endpoint:    ngfTracing.Endpoint,
			ServiceName: gwNsName + ":ngf",
		}
		if ngfTracing.Enable != nil {
			tracing.Enabled = *ngfTracing.Enable
		} else {
			tracing.Enabled = false
		}
		if ngfTracing.BatchCount != nil {
			tracing.BatchCount = *ngfTracing.BatchCount
		} else {
			tracing.BatchCount = 4
		}
		if ngfTracing.BatchSize != nil {
			tracing.BatchSize = *ngfTracing.BatchSize
		} else {
			tracing.BatchSize = 512
		}
		if ngfTracing.Interval != nil {
			tracing.Interval = *ngfTracing.Interval
		} else {
			tracing.Interval = "5s"
		}
	}

	return tracing, valErr
}
