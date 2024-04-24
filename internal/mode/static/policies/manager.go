package policies

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GenerateFunc generates config as []byte for an NGF Policy.
type GenerateFunc func(policy Policy) []byte

// Validator validates an NGF Policy.
type Validator interface {
	// Validate validates an NGF Policy.
	Validate(policy Policy) error
	// Conflicts returns true if the two Policies conflict.
	Conflicts(a, b Policy) bool
}

// Manager manages the validators and generators for NGF Policies.
type Manager struct {
	validators     map[schema.GroupVersionKind]Validator
	generators     map[schema.GroupVersionKind]GenerateFunc
	mustExtractGVK func(client.Object) schema.GroupVersionKind
}

// ManagerConfig contains the config to register a Policy with the Manager.
type ManagerConfig struct {
	// GVK is the GroupVersionKind of the Policy.
	GVK schema.GroupVersionKind
	// Validator is the Validator for the Policy.
	Validator Validator
	// Generator is the GenerateFunc for the Policy.
	Generator GenerateFunc
}

// NewManager returns a new Manager.
// Implements dataplane.PolicyConfigGenerator and validation.PolicyValidator.
func NewManager(
	mustExtractGVK func(client.Object) schema.GroupVersionKind,
	configs ...ManagerConfig,
) *Manager {
	v := &Manager{
		validators:     make(map[schema.GroupVersionKind]Validator),
		generators:     make(map[schema.GroupVersionKind]GenerateFunc),
		mustExtractGVK: mustExtractGVK,
	}

	for _, cfg := range configs {
		v.validators[cfg.GVK] = cfg.Validator
		v.generators[cfg.GVK] = cfg.Generator
	}

	return v
}

// Generate generates config for the policy as a byte array.
func (m *Manager) Generate(policy Policy) []byte {
	gvk := m.mustExtractGVK(policy)

	generate, ok := m.generators[gvk]
	if !ok {
		panic(fmt.Sprintf("no generate function registered for policy %T", policy))
	}

	return generate(policy)
}

// Validate validates the policy.
func (m *Manager) Validate(policy Policy) error {
	gvk := m.mustExtractGVK(policy)

	validator, ok := m.validators[gvk]
	if !ok {
		panic(fmt.Sprintf("no validator registered for policy %T", policy))
	}

	return validator.Validate(policy)
}

// Conflicts returns true if the policies conflict.
func (m *Manager) Conflicts(polA, polB Policy) bool {
	gvk := m.mustExtractGVK(polA)

	validator, ok := m.validators[gvk]
	if !ok {
		panic(fmt.Sprintf("no validator registered for policy %T", polA))
	}

	return validator.Conflicts(polA, polB)
}
