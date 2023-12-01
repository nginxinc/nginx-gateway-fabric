package graph

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// Namespace represents a Namespace resource.
type Namespace struct {
	// Source holds the actual Namespace resource. Can be nil if the Namespace does not exist.
	Source *v1.Namespace
}

type namespaceHolder struct {
	clusterNamespaces    map[types.NamespacedName]*v1.Namespace
	referencedNamespaces map[types.NamespacedName]*Namespace
}

func newNamespaceHolder(namespaces map[types.NamespacedName]*v1.Namespace) namespaceHolder {
	return namespaceHolder{
		clusterNamespaces:    namespaces,
		referencedNamespaces: make(map[types.NamespacedName]*Namespace),
	}
}

func buildReferencedNamespaces(namespaceHolder namespaceHolder, gw *Gateway) {
	for name, ns := range namespaceHolder.clusterNamespaces {
		if checkNamespace(ns, gw) {
			namespaceHolder.referencedNamespaces[name] = &Namespace{Source: ns}
		}
	}
}

// checkNamespaces returns a boolean that represents whether a given Namespace resource has a label
// that matches any of the Gateway Listener's label selector.
func checkNamespace(ns *v1.Namespace, gw *Gateway) bool {
	if gw == nil || ns == nil {
		return false
	}
	nsLabels := ns.GetLabels()
	for _, listener := range gw.Listeners {
		if listener.AllowedRouteLabelSelector == nil {
			// Can have listeners with AllowedRouteLabelSelector not set.
			continue
		}
		if listener.AllowedRouteLabelSelector.Matches(labels.Set(nsLabels)) {
			return true
		}
	}
	return false
}

func (r *namespaceHolder) getResolvedNamespaces() map[types.NamespacedName]*Namespace {
	if len(r.referencedNamespaces) == 0 {
		return nil
	}
	return r.referencedNamespaces
}
