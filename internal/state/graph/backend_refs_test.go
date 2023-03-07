package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func getNormalRef() v1beta1.BackendRef {
	return v1beta1.BackendRef{
		BackendObjectReference: v1beta1.BackendObjectReference{
			Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Service")),
			Name:      "service1",
			Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
			Port:      (*v1beta1.PortNumber)(helpers.GetInt32Pointer(80)),
		},
		Weight: helpers.GetInt32Pointer(1),
	}
}

func getModifiedRef(mod func(ref v1beta1.BackendRef) v1beta1.BackendRef) v1beta1.BackendRef {
	return mod(getNormalRef())
}

func TestValidateHTTPBackendRef(t *testing.T) {
	tests := []struct {
		expectedCondition conditions.Condition
		name              string
		ref               v1beta1.HTTPBackendRef
		expectedValid     bool
	}{
		{
			name: "normal case",
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getNormalRef(),
				Filters:    nil,
			},
			expectedValid: true,
		},
		{
			name: "filters not supported",
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getNormalRef(),
				Filters: []v1beta1.HTTPRouteFilter{
					{
						Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
					},
				},
			},
			expectedValid: false,
			expectedCondition: conditions.NewRouteBackendRefUnsupportedValue(
				"test.filters: Too many: 1: must have at most 0 items",
			),
		},
		{
			name: "invalid base ref",
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
					backend.Kind = helpers.GetPointer[v1beta1.Kind]("NotService")
					return backend
				}),
			},
			expectedValid: false,
			expectedCondition: conditions.NewRouteBackendRefInvalidKind(
				`test.kind: Unsupported value: "NotService": supported values: "Service"`,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			valid, cond := validateHTTPBackendRef(test.ref, "test", field.NewPath("test"))

			g.Expect(valid).To(Equal(test.expectedValid))
			g.Expect(cond).To(Equal(test.expectedCondition))
		})
	}
}

func TestValidateBackendRef(t *testing.T) {
	tests := []struct {
		ref               v1beta1.BackendRef
		expectedCondition conditions.Condition
		name              string
		expectedValid     bool
	}{
		{
			name:          "normal case",
			ref:           getNormalRef(),
			expectedValid: true,
		},
		{
			name: "normal case with implicit namespace",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Namespace = nil
				return backend
			}),
			expectedValid: true,
		},
		{
			name: "normal case with implicit kind Service",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Kind = nil
				return backend
			}),
			expectedValid: true,
		},
		{
			name: "invalid group",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Group = helpers.GetPointer[v1beta1.Group]("invalid")
				return backend
			}),
			expectedValid: false,
			expectedCondition: conditions.NewRouteBackendRefUnsupportedValue(
				`test.group: Unsupported value: "invalid": supported values: "core", ""`,
			),
		},
		{
			name: "not a service kind",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Kind = (*v1beta1.Kind)(helpers.GetStringPointer("NotService"))
				return backend
			}),
			expectedValid: false,
			expectedCondition: conditions.NewRouteBackendRefInvalidKind(
				`test.kind: Unsupported value: "NotService": supported values: "Service"`,
			),
		},
		{
			name: "invalid namespace",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Namespace = (*v1beta1.Namespace)(helpers.GetStringPointer("invalid"))
				return backend
			}),
			expectedValid: false,
			expectedCondition: conditions.NewRouteBackendRefRefNotPermitted(
				`test.namespace: Invalid value: "invalid": cross-namespace routing is not permitted`,
			),
		},
		{
			name: "invalid weight",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Weight = helpers.GetPointer[int32](-1)
				return backend
			}),
			expectedValid: false,
			expectedCondition: conditions.NewRouteBackendRefUnsupportedValue(
				"test.weight: Invalid value: -1: must be in the range [0, 1000000]",
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			valid, cond := validateBackendRef(test.ref, "test", field.NewPath("test"))

			g.Expect(valid).To(Equal(test.expectedValid))
			g.Expect(cond).To(Equal(test.expectedCondition))
		})
	}
}

func TestValidateWeight(t *testing.T) {
	validWeights := []int32{0, 1, 1000000}
	invalidWeights := []int32{-1, 1000001}

	g := NewGomegaWithT(t)

	for _, w := range validWeights {
		err := validateWeight(w)
		g.Expect(err).ToNot(HaveOccurred(), "Expected weight %d to be valid", w)
	}
	for _, w := range invalidWeights {
		err := validateWeight(w)
		g.Expect(err).To(HaveOccurred(), "Expected weight %d to be invalid", w)
	}
}

func TestGetServiceAndPortFromRef(t *testing.T) {
	svc1 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service1",
			Namespace: "test",
		},
	}

	svc2 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service2",
			Namespace: "test",
		},
	}

	tests := []struct {
		ref        v1beta1.BackendRef
		expService *v1.Service
		name       string
		expPort    int32
		expErr     bool
	}{
		{
			name:       "normal case",
			ref:        getNormalRef(),
			expService: svc1,
			expPort:    80,
		},
		{
			name: "service does not exist",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Name = "does-not-exist"
				return backend
			}),
			expErr: true,
		},
	}

	services := map[types.NamespacedName]*v1.Service{
		{Namespace: "test", Name: "service1"}: svc1,
		{Namespace: "test", Name: "service2"}: svc2,
	}

	refPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			svc, port, err := getServiceAndPortFromRef(test.ref, "test", services, refPath)

			g.Expect(err != nil).To(Equal(test.expErr))
			g.Expect(svc).To(Equal(test.expService))
			g.Expect(port).To(Equal(test.expPort))
		})
	}
}

func TestAddBackendGroupsToRouteTest(t *testing.T) {
	createRoute := func(name string, kind v1beta1.Kind, refsPerBackend int, serviceNames ...string) *v1beta1.HTTPRoute {
		hr := &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
		}

		createHTTPBackendRef := func(svcName string, port v1beta1.PortNumber, weight *int32) v1beta1.HTTPBackendRef {
			return v1beta1.HTTPBackendRef{
				BackendRef: v1beta1.BackendRef{
					BackendObjectReference: v1beta1.BackendObjectReference{
						Kind:      helpers.GetPointer(kind),
						Name:      v1beta1.ObjectName(svcName),
						Namespace: helpers.GetPointer[v1beta1.Namespace]("test"),
						Port:      helpers.GetPointer(port),
					},
					Weight: weight,
				},
			}
		}

		hr.Spec.Rules = make([]v1beta1.HTTPRouteRule, len(serviceNames))

		for idx, svcName := range serviceNames {
			refs := []v1beta1.HTTPBackendRef{
				createHTTPBackendRef(svcName, 80, nil),
			}
			if refsPerBackend == 2 {
				refs = append(refs, createHTTPBackendRef(svcName, 81, helpers.GetPointer[int32](5)))
			}
			if refsPerBackend != 1 && refsPerBackend != 2 {
				panic("invalid refsPerBackend")
			}

			hr.Spec.Rules[idx] = v1beta1.HTTPRouteRule{
				BackendRefs: refs,
			}
		}
		return hr
	}

	const (
		allValid   = true
		allInvalid = false
	)

	createRules := func(hr *v1beta1.HTTPRoute, validMatches, validFilters bool) []Rule {
		rules := make([]Rule, len(hr.Spec.Rules))
		for i := range rules {
			rules[i].ValidMatches = validMatches
			rules[i].ValidFilters = validFilters
			rules[i].BackendGroup = BackendGroup{
				Source:  client.ObjectKeyFromObject(hr),
				RuleIdx: i,
			}
		}
		return rules
	}

	const sectionName = "test"
	sectionNameRefs := map[string]ParentRef{
		sectionName: {
			Idx:     0,
			Gateway: types.NamespacedName{Namespace: "test", Name: "gateway"},
		},
	}

	hrWithOneBackend := createRoute("hr1", "Service", 1, "svc1")
	hrWithTwoBackends := createRoute("hr2", "Service", 2, "svc1")
	hrWithInvalidRule := createRoute("hr3", "NotService", 1, "svc1")
	hrWithZeroBackendRefs := createRoute("hr4", "Service", 1, "svc1")
	hrWithZeroBackendRefs.Spec.Rules[0].BackendRefs = nil

	svc1 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "svc1"}}

	services := map[types.NamespacedName]*v1.Service{
		{Namespace: "test", Name: "svc1"}: svc1,
	}

	tests := []struct {
		name                  string
		route                 *Route
		expectedBackendGroups []BackendGroup
		expectedConditions    []conditions.Condition
	}{
		{
			route: &Route{
				Source:          hrWithOneBackend,
				SectionNameRefs: sectionNameRefs,
				Valid:           true,
				Rules:           createRules(hrWithOneBackend, allValid, allValid),
			},
			expectedBackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hrWithOneBackend),
					RuleIdx: 0,
					Backends: []BackendRef{
						{
							Name:   "test_svc1_80",
							Svc:    svc1,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
					},
				},
			},
			expectedConditions: nil,
			name:               "normal case with one rule with one backend",
		},
		{
			route: &Route{
				Source:          hrWithTwoBackends,
				SectionNameRefs: sectionNameRefs,
				Valid:           true,
				Rules:           createRules(hrWithTwoBackends, allValid, allValid),
			},
			expectedBackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hrWithTwoBackends),
					RuleIdx: 0,
					Backends: []BackendRef{
						{
							Name:   "test_svc1_80",
							Svc:    svc1,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
						{
							Name:   "test_svc1_81",
							Svc:    svc1,
							Port:   81,
							Valid:  true,
							Weight: 5,
						},
					},
				},
			},
			expectedConditions: nil,
			name:               "normal case with one rule with two backends",
		},
		{
			route: &Route{
				Source:          hrWithOneBackend,
				SectionNameRefs: sectionNameRefs,
				Valid:           false,
			},
			expectedBackendGroups: nil,
			expectedConditions:    nil,
			name:                  "invalid route",
		},
		{
			route: &Route{
				Source:          hrWithOneBackend,
				SectionNameRefs: sectionNameRefs,
				Valid:           true,
				Rules:           createRules(hrWithOneBackend, allInvalid, allValid),
			},
			expectedBackendGroups: nil,
			expectedConditions:    nil,
			name:                  "invalid matches",
		},
		{
			route: &Route{
				Source:          hrWithOneBackend,
				SectionNameRefs: sectionNameRefs,
				Valid:           true,
				Rules:           createRules(hrWithOneBackend, allValid, allInvalid),
			},
			expectedBackendGroups: nil,
			expectedConditions:    nil,
			name:                  "invalid filters",
		},
		{
			route: &Route{
				Source:          hrWithInvalidRule,
				SectionNameRefs: sectionNameRefs,
				Valid:           true,
				Rules:           createRules(hrWithInvalidRule, allValid, allValid),
			},
			expectedBackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hrWithInvalidRule),
					RuleIdx: 0,
					Backends: []BackendRef{
						{
							Weight: 1,
						},
					},
				},
			},
			expectedConditions: []conditions.Condition{
				conditions.NewRouteBackendRefInvalidKind(
					`spec.rules[0].backendRefs[0].kind: Unsupported value: "NotService": supported values: "Service"`,
				),
			},
			name: "invalid backendRef",
		},
		{
			route: &Route{
				Source:          hrWithZeroBackendRefs,
				SectionNameRefs: sectionNameRefs,
				Valid:           true,
				Rules:           createRules(hrWithZeroBackendRefs, allValid, allValid),
			},
			expectedBackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hrWithZeroBackendRefs),
					RuleIdx: 0,
				},
			},
			expectedConditions: nil,
			name:               "zero backendRefs",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			addBackendGroupsToRoute(test.route, services)

			g.Expect(helpers.Diff(test.expectedBackendGroups, test.route.GetAllBackendGroups())).To(BeEmpty())
			g.Expect(test.route.GetAllConditionsForSectionName(sectionName)).To(Equal(test.expectedConditions))
		})
	}
}

func TestCreateBackend(t *testing.T) {
	svc1 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "service1"}}

	tests := []struct {
		name               string
		ref                v1beta1.HTTPBackendRef
		expectedConditions []conditions.Condition
		expectedBackend    BackendRef
	}{
		{
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getNormalRef(),
			},
			expectedBackend: BackendRef{
				Svc:    svc1,
				Name:   "test_service1_80",
				Port:   80,
				Weight: 1,
				Valid:  true,
			},
			name: "normal case",
		},
		{
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
					backend.Weight = nil
					return backend
				}),
			},
			expectedBackend: BackendRef{
				Svc:    svc1,
				Name:   "test_service1_80",
				Port:   80,
				Weight: 1,
				Valid:  true,
			},
			name: "normal with nil weight",
		},
		{
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
					backend.Weight = helpers.GetPointer[int32](-1)
					return backend
				}),
			},
			expectedBackend: BackendRef{
				Svc:    nil,
				Name:   "",
				Port:   0,
				Weight: 0,
				Valid:  false,
			},
			expectedConditions: []conditions.Condition{
				conditions.NewRouteBackendRefUnsupportedValue("test.weight: Invalid value: -1: must be in the range [0, 1000000]"),
			},
			name: "invalid weight",
		},
		{
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
					backend.Kind = helpers.GetPointer[v1beta1.Kind]("NotService")
					return backend
				}),
			},
			expectedBackend: BackendRef{
				Svc:    nil,
				Name:   "",
				Port:   0,
				Weight: 1,
				Valid:  false,
			},
			expectedConditions: []conditions.Condition{
				conditions.NewRouteBackendRefInvalidKind(`test.kind: Unsupported value: "NotService": supported values: "Service"`),
			},
			name: "invalid kind",
		},
		{
			ref: v1beta1.HTTPBackendRef{
				BackendRef: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
					backend.Name = "not-exist"
					return backend
				}),
			},
			expectedBackend: BackendRef{
				Svc:    nil,
				Name:   "",
				Port:   0,
				Weight: 1,
				Valid:  false,
			},
			expectedConditions: []conditions.Condition{
				conditions.NewRouteBackendRefRefBackendNotFound(`test.name: Not found: "not-exist"`),
			},
			name: "service doesn't exist",
		},
	}

	services := map[types.NamespacedName]*v1.Service{
		client.ObjectKeyFromObject(svc1): svc1,
	}
	sourceNamespace := "test"

	refPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			backend, conds := createBackend(test.ref, sourceNamespace, services, refPath)

			g.Expect(helpers.Diff(test.expectedBackend, backend)).To(BeEmpty())
			g.Expect(conds).To(Equal(test.expectedConditions))
		})
	}
}
