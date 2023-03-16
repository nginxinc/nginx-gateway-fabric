package graph

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// BackendGroup represents a group of backends for a rule in an HTTPRoute.
type BackendGroup struct {
	Source   types.NamespacedName
	Backends []BackendRef
	RuleIdx  int
}

// BackendRef is an internal representation of a backendRef in an HTTPRoute.
type BackendRef struct {
	Svc    *v1.Service
	Name   string
	Port   int32
	Weight int32
	Valid  bool
}

// GroupName returns the name of the backend group.
// This name must be unique across all HTTPRoutes and all rules within the same HTTPRoute.
// The RuleIdx is used to make the name unique across all rules within the same HTTPRoute.
// The RuleIdx may change for a given rule if an update is made to the HTTPRoute, but it will always match the index
// of the rule in the stored HTTPRoute.
func (bg *BackendGroup) GroupName() string {
	return fmt.Sprintf("%s__%s_rule%d", bg.Source.Namespace, bg.Source.Name, bg.RuleIdx)
}
