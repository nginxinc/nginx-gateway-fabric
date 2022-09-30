package state

import (
	"fmt"
	"sort"
)

// InvalidBackendRef is the upstream name for a backend ref that is invalid.
// Invalid in this case means that a Kubernetes Service cannot be extracted from it.
const InvalidBackendRef = "invalid_backend_ref"

func generateUpstreamName(service backendService) string {
	if service.Name == "" {
		return InvalidBackendRef
	}
	return fmt.Sprintf("%s_%s_%d", service.Namespace, service.Name, service.Port)
}

func buildUpstreams(backends map[backendService]backend) []Upstream {
	upstreams := make([]Upstream, 0, len(backends))

	for svc, b := range backends {
		upstreams = append(upstreams, Upstream{
			Name:      generateUpstreamName(svc),
			Endpoints: b.Endpoints,
		})
	}

	// sort upstreams for test-ability
	sort.Slice(upstreams, func(i, j int) bool {
		return upstreams[i].Name < upstreams[j].Name
	})

	return upstreams
}
