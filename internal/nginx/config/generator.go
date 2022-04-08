package config

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
const nginx502Server = "unix:/var/lib/nginx/nginx-502-server.sock"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
type Generator interface {
	// GenerateForHost generates configuration for a host.
	GenerateForHost(host state.Host) ([]byte, Warnings)
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

func (g *GeneratorImpl) GenerateForHost(host state.Host) ([]byte, Warnings) {
	server, warnings := generate(host, g.serviceStore)
	return g.executor.ExecuteForServer(server), warnings
}

func generate(host state.Host, serviceStore state.ServiceStore) (server, Warnings) {
	warnings := newWarnings()

	locs := make([]location, 0, len(host.PathRouteGroups)) // FIXME(pleshakov): expand with g.Routes

	for _, g := range host.PathRouteGroups {
		// number of routes in a group is always at least 1
		// otherwise, it is a bug in the state.Configuration code, so it is OK to panic here
		r := g.Routes[0] // FIXME(pleshakov): for now, we only handle the first route in case there are multiple routes
		address, err := getBackendAddress(r.Source.Spec.Rules[r.RuleIdx].BackendRefs, r.Source.Namespace, serviceStore)
		if err != nil {
			warnings.AddWarning(r.Source, err.Error())
		}

		loc := location{
			Path:      g.Path,
			ProxyPass: generateProxyPass(address),
		}
		locs = append(locs, loc)
	}

	return server{
		ServerName: host.Value,
		Locations:  locs,
	}, warnings
}

func generateProxyPass(address string) string {
	if address == "" {
		return "http://" + nginx502Server
	}
	return "http://" + address
}

func getBackendAddress(refs []v1alpha2.HTTPBackendRef, parentNS string, serviceStore state.ServiceStore) (string, error) {
	if len(refs) == 0 {
		return "", errors.New("empty backend refs")
	}

	// FIXME(pleshakov): for now, we only support a single backend reference
	ref := refs[0].BackendRef

	if ref.Kind != nil && *ref.Kind != "Service" {
		return "", fmt.Errorf("unsupported kind %s", *ref.Kind)
	}

	ns := parentNS
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	address, err := serviceStore.Resolve(types.NamespacedName{Namespace: ns, Name: string(ref.Name)})
	if err != nil {
		return "", fmt.Errorf("service %s/%s cannot be resolved: %w", ns, ref.Name, err)
	}

	if ref.Port == nil {
		return "", errors.New("port is nil")
	}

	return fmt.Sprintf("%s:%d", address, *ref.Port), nil
}
