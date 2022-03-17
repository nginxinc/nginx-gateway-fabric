package config

import (
	"fmt"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
const nginx502Server = "unix:/var/lib/nginx/nginx-502-server.sock"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
type Generator interface {
	// GenerateForHost generates configuration for a host.
	GenerateForHost(host state.Host) []byte
}

// GeneratorImpl is an implementation of Generator
type GeneratorImpl struct {
	executor     *templateExecutor
	serviceStore state.ServiceStore
}

// NewGeneratorImpl creates a new GeneratorImpl.
func NewGeneratorImpl(serviceStore state.ServiceStore) *GeneratorImpl {
	return &GeneratorImpl{
		executor:     newTemplateExecutor(),
		serviceStore: serviceStore,
	}
}

func (g *GeneratorImpl) GenerateForHost(host state.Host) []byte {
	server := generate(host, g.serviceStore)
	return g.executor.ExecuteForServer(server)
}

func generate(host state.Host, serviceStore state.ServiceStore) server {
	var locs []location

	for _, g := range host.PathRouteGroups {
		// number of routes in a group is always at least 1
		// otherwise, it is a bug in the state.Configuration code, so it is OK to panic here
		r := g.Routes[0] // TO-DO: for now, we only handle the first route in case there are multiple routes
		address := getBackendAddress(r.Source.Spec.Rules[r.RuleIdx].BackendRefs, r.Source.Namespace, serviceStore)

		loc := location{
			Path:      g.Path,
			ProxyPass: generateProxyPass(address),
		}
		locs = append(locs, loc)
	}

	return server{
		ServerName: host.Value,
		Locations:  locs,
	}
}

func generateProxyPass(address string) string {
	if address == "" {
		return "http://" + nginx502Server
	}
	return "http://" + address
}

func getBackendAddress(refs []v1alpha2.HTTPBackendRef, parentNS string, serviceStore state.ServiceStore) string {
	// TO-DO: make sure the warnings are generated and reported to the user fot the edge cases
	if len(refs) == 0 {
		return ""
	}

	// TO-DO: for now, we only support a single backend reference
	ref := refs[0].BackendRef

	if ref.Kind != nil && *ref.Kind != "Service" {
		return ""
	}

	ns := parentNS
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	address, err := serviceStore.Resolve(types.NamespacedName{Namespace: ns, Name: string(ref.Name)})
	if err != nil {
		return ""
	}

	if ref.Port == nil {
		return ""
	}

	return fmt.Sprintf("%s:%d", address, *ref.Port)
}
