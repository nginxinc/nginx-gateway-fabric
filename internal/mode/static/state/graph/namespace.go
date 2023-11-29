package graph

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

type Namespace struct {
	Source *v1.Namespace
}

type namespaceResolver struct {
	clusterNamespaces  map[types.NamespacedName]*v1.Namespace
	resolvedNamespaces map[types.NamespacedName]*Namespace
}

func newNamespaceResolver(namespaces map[types.NamespacedName]*v1.Namespace) namespaceResolver {
	return namespaceResolver{
		clusterNamespaces:  namespaces,
		resolvedNamespaces: make(map[types.NamespacedName]*Namespace),
	}
}

func resolveNamespaces(namespaceResolver namespaceResolver, gw *Gateway) {
	for name, ns := range namespaceResolver.clusterNamespaces {
		if checkNamespace(ns, gw) {
			namespaceResolver.resolvedNamespaces[name] = &Namespace{Source: ns}
		}
	}
}

// checkNamespaces returns a boolean that represents whether a given Namespace resource has a label
// that matches any of the Gateway Listener's label selector.
func checkNamespace(ns *v1.Namespace, gw *Gateway) bool {
	nsLabels := ns.GetLabels()
	if gw == nil {
		return false
	}
	for _, listener := range gw.Listeners {
		if listener.AllowedRouteLabelSelector == nil {
			return false
		}
		if listener.AllowedRouteLabelSelector.Matches(labels.Set(nsLabels)) {
			return true
		}
	}
	return false
}

func (r *namespaceResolver) getResolvedNamespaces() map[types.NamespacedName]*Namespace {
	if len(r.resolvedNamespaces) == 0 {
		return nil
	}
	return r.resolvedNamespaces
}
