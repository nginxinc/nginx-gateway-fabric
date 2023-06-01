package state_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/relationship"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/relationship/relationshipfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets/secretsfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation/validationfakes"
)

const (
	controllerName  = "my.controller"
	gcName          = "test-class"
	certificatePath = "path/to/cert"
)

func createRoute(
	name string,
	gateway string,
	hostname string,
	backendRefs ...v1beta1.HTTPBackendRef,
) *v1beta1.HTTPRoute {
	return &v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  "test",
			Name:       name,
			Generation: 1,
		},
		Spec: v1beta1.HTTPRouteSpec{
			CommonRouteSpec: v1beta1.CommonRouteSpec{
				ParentRefs: []v1beta1.ParentReference{
					{
						Namespace: (*v1beta1.Namespace)(helpers.GetPointer("test")),
						Name:      v1beta1.ObjectName(gateway),
						SectionName: (*v1beta1.SectionName)(
							helpers.GetPointer("listener-80-1"),
						),
					},
					{
						Namespace: (*v1beta1.Namespace)(helpers.GetPointer("test")),
						Name:      v1beta1.ObjectName(gateway),
						SectionName: (*v1beta1.SectionName)(
							helpers.GetPointer("listener-443-1"),
						),
					},
				},
			},
			Hostnames: []v1beta1.Hostname{
				v1beta1.Hostname(hostname),
			},
			Rules: []v1beta1.HTTPRouteRule{
				{
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
								Value: helpers.GetPointer("/"),
							},
						},
					},
					BackendRefs: backendRefs,
				},
			},
		},
	}
}

func createGateway(name string) *v1beta1.Gateway {
	return &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  "test",
			Name:       name,
			Generation: 1,
		},
		Spec: v1beta1.GatewaySpec{
			GatewayClassName: gcName,
			Listeners: []v1beta1.Listener{
				{
					Name:     "listener-80-1",
					Hostname: nil,
					Port:     80,
					Protocol: v1beta1.HTTPProtocolType,
				},
			},
		},
	}
}

func createGatewayWithTLSListener(name string) *v1beta1.Gateway {
	gw := createGateway(name)

	l := v1beta1.Listener{
		Name:     "listener-443-1",
		Hostname: nil,
		Port:     443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS: &v1beta1.GatewayTLSConfig{
			Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
					Name:      "secret",
					Namespace: (*v1beta1.Namespace)(helpers.GetPointer("test")),
				},
			},
		},
	}
	gw.Spec.Listeners = append(gw.Spec.Listeners, l)

	return gw
}

func createRouteWithMultipleRules(
	name, gateway, hostname string,
	rules []v1beta1.HTTPRouteRule,
) *v1beta1.HTTPRoute {
	hr := createRoute(name, gateway, hostname)
	hr.Spec.Rules = rules

	return hr
}

func createHTTPRule(path string, backendRefs ...v1beta1.HTTPBackendRef) v1beta1.HTTPRouteRule {
	return v1beta1.HTTPRouteRule{
		Matches: []v1beta1.HTTPRouteMatch{
			{
				Path: &v1beta1.HTTPPathMatch{
					Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
					Value: &path,
				},
			},
		},
		BackendRefs: backendRefs,
	}
}

func createBackendRef(
	kind *v1beta1.Kind,
	name v1beta1.ObjectName,
	namespace *v1beta1.Namespace,
) v1beta1.HTTPBackendRef {
	return v1beta1.HTTPBackendRef{
		BackendRef: v1beta1.BackendRef{
			BackendObjectReference: v1beta1.BackendObjectReference{
				Kind:      kind,
				Name:      name,
				Namespace: namespace,
				Port:      helpers.GetPointer[v1beta1.PortNumber](80),
			},
		},
	}
}

func createAlwaysValidValidators() validation.Validators {
	http := &validationfakes.FakeHTTPFieldsValidator{}

	return validation.Validators{
		HTTPFieldsValidator: http,
	}
}

func createScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(discoveryV1.AddToScheme(scheme))

	return scheme
}

var _ = Describe("ChangeProcessor", func() {
	format.MaxLength = 0
	Describe("Normal cases of processing changes", func() {
		var (
			gc = &v1beta1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       gcName,
					Generation: 1,
				},
				Spec: v1beta1.GatewayClassSpec{
					ControllerName: controllerName,
				},
			}
			processor           state.ChangeProcessor
			fakeSecretMemoryMgr *secretsfakes.FakeSecretDiskMemoryManager
		)

		BeforeEach(OncePerOrdered, func() {
			fakeSecretMemoryMgr = &secretsfakes.FakeSecretDiskMemoryManager{}

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      controllerName,
				GatewayClassName:     gcName,
				SecretMemoryManager:  fakeSecretMemoryMgr,
				RelationshipCapturer: relationship.NewCapturerImpl(),
				Logger:               zap.New(),
				Validators:           createAlwaysValidValidators(),
				Scheme:               createScheme(),
			})

			fakeSecretMemoryMgr.RequestReturns(certificatePath, nil)
		})

		Describe("Process gateway resources", Ordered, func() {
			var (
				gcUpdated                *v1beta1.GatewayClass
				hr1, hr1Updated, hr2     *v1beta1.HTTPRoute
				gw1, gw1Updated, gw2     *v1beta1.Gateway
				expGraph                 *graph.Graph
				expRouteHR1, expRouteHR2 *graph.Route
			)
			BeforeAll(func() {
				gcUpdated = gc.DeepCopy()
				gcUpdated.Generation++

				hr1 = createRoute("hr-1", "gateway-1", "foo.example.com")

				hr1Updated = hr1.DeepCopy()
				hr1Updated.Generation++

				hr2 = createRoute("hr-2", "gateway-2", "bar.example.com")

				gw1 = createGatewayWithTLSListener("gateway-1")

				gw1Updated = gw1.DeepCopy()
				gw1Updated.Generation++

				gw2 = createGatewayWithTLSListener("gateway-2")
			})
			BeforeEach(func() {
				expRouteHR1 = &graph.Route{
					Source: hr1,
					ParentRefs: []graph.ParentRef{
						{
							Attachment: &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{"listener-80-1": {"foo.example.com"}},
								Attached:          true,
							},
							Gateway: types.NamespacedName{Namespace: "test", Name: "gateway-1"},
						},
						{
							Attachment: &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{"listener-443-1": {"foo.example.com"}},
								Attached:          true,
							},
							Gateway: types.NamespacedName{Namespace: "test", Name: "gateway-1"},
							Idx:     1,
						},
					},
					Rules: []graph.Rule{{ValidMatches: true, ValidFilters: true}},
					Valid: true,
				}

				expRouteHR2 = &graph.Route{
					Source: hr2,
					ParentRefs: []graph.ParentRef{
						{
							Attachment: &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{"listener-80-1": {"bar.example.com"}},
								Attached:          true,
							},
							Gateway: types.NamespacedName{Namespace: "test", Name: "gateway-2"},
						},
						{
							Attachment: &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{"listener-443-1": {"bar.example.com"}},
								Attached:          true,
							},
							Gateway: types.NamespacedName{Namespace: "test", Name: "gateway-2"},
							Idx:     1,
						},
					},
					Rules: []graph.Rule{{ValidMatches: true, ValidFilters: true}},
					Valid: true,
				}

				// This is the base case expected graph. Tests will manipulate this to add or remove elements
				// to fit the expected output of the input under test.
				expGraph = &graph.Graph{
					GatewayClass: &graph.GatewayClass{
						Source: gc,
						Valid:  true,
					},
					Gateway: &graph.Gateway{
						Source: gw1,
						Listeners: map[string]*graph.Listener{
							"listener-80-1": {
								Source: gw1.Spec.Listeners[0],
								Valid:  true,
								Routes: map[types.NamespacedName]*graph.Route{
									{Namespace: "test", Name: "hr-1"}: expRouteHR1,
								},
							},
							"listener-443-1": {
								Source: gw1.Spec.Listeners[1],
								Valid:  true,
								Routes: map[types.NamespacedName]*graph.Route{
									{Namespace: "test", Name: "hr-1"}: expRouteHR1,
								},
								SecretPath: "path/to/cert",
							},
						},
						Valid: true,
					},
					IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{},
					Routes: map[types.NamespacedName]*graph.Route{
						{Namespace: "test", Name: "hr-1"}: expRouteHR1,
					},
				}
			})

			When("no upsert has occurred", func() {
				It("returns empty graph", func() {
					changed, graphCfg := processor.Process()
					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())
				})
			})
			When("GatewayClass doesn't exist", func() {
				When("Gateways don't exist", func() {
					When("the first HTTPRoute is upserted", func() {
						It("returns empty graph", func() {
							processor.CaptureUpsertChange(hr1)

							changed, graphCfg := processor.Process()
							Expect(changed).To(BeTrue())
							Expect(graphCfg).To(Equal(&graph.Graph{}))
						})
					})
					When("the first Gateway is upserted", func() {
						It("returns populated graph", func() {
							processor.CaptureUpsertChange(gw1)

							expGraph.GatewayClass = nil

							expGraph.Gateway.Conditions = conditions.NewGatewayInvalid("GatewayClass doesn't exist")
							expGraph.Gateway.Valid = false
							expGraph.Gateway.Listeners = nil

							hrName := types.NamespacedName{Namespace: "test", Name: "hr-1"}
							expGraph.Routes[hrName].ParentRefs[0].Attachment = &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{},
								FailedCondition:   conditions.NewRouteInvalidGateway(),
							}
							expGraph.Routes[hrName].ParentRefs[1].Attachment = &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{},
								FailedCondition:   conditions.NewRouteInvalidGateway(),
							}

							changed, graphCfg := processor.Process()
							Expect(changed).To(BeTrue())
							Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
						})
					})
				})
			})
			When("the GatewayClass is upserted", func() {
				It("returns updated graph", func() {
					processor.CaptureUpsertChange(gc)

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the first HTTPRoute without a generation changed is processed", func() {
				It("returns empty graph", func() {
					hr1UpdatedSameGen := hr1.DeepCopy()
					// hr1UpdatedSameGen.Generation has not been changed
					processor.CaptureUpsertChange(hr1UpdatedSameGen)

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())
				})
			})
			When("the first HTTPRoute update with a generation changed is processed", func() {
				It("returns populated graph", func() {
					processor.CaptureUpsertChange(hr1Updated)

					hrName := types.NamespacedName{Namespace: "test", Name: "hr-1"}
					expGraph.Gateway.Listeners["listener-443-1"].Routes[hrName].Source.Generation = hr1Updated.Generation
					expGraph.Gateway.Listeners["listener-80-1"].Routes[hrName].Source.Generation = hr1Updated.Generation

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				},
				)
			})
			When("the first Gateway update without generation changed is processed", func() {
				It("returns empty graph", func() {
					gwUpdatedSameGen := gw1.DeepCopy()
					// gwUpdatedSameGen.Generation has not been changed
					processor.CaptureUpsertChange(gwUpdatedSameGen)

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())
				})
			})
			When("the first Gateway update with a generation changed is processed", func() {
				It("returns populated graph", func() {
					processor.CaptureUpsertChange(gw1Updated)

					expGraph.Gateway.Source.Generation = gw1Updated.Generation

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the GatewayClass update without generation change is processed", func() {
				It("returns empty graph", func() {
					gcUpdatedSameGen := gc.DeepCopy()
					// gcUpdatedSameGen.Generation has not been changed
					processor.CaptureUpsertChange(gcUpdatedSameGen)

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())
				})
			})
			When("the GatewayClass update with generation change is processed", func() {
				It("returns populated graph", func() {
					processor.CaptureUpsertChange(gcUpdated)

					expGraph.GatewayClass.Source.Generation = gcUpdated.Generation

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("no changes are captured", func() {
				It("returns empty graph", func() {
					changed, graphCfg := processor.Process()

					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())
				})
			})
			When("the second Gateway is upserted", func() {
				It("returns populated graph using first gateway", func() {
					processor.CaptureUpsertChange(gw2)

					expGraph.IgnoredGateways = map[types.NamespacedName]*v1beta1.Gateway{
						{Namespace: "test", Name: "gateway-2"}: gw2,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the second HTTPRoute is upserted", func() {
				It("returns populated graph", func() {
					processor.CaptureUpsertChange(hr2)

					hrName := types.NamespacedName{Namespace: "test", Name: "hr-2"}

					expGraph.IgnoredGateways = map[types.NamespacedName]*v1beta1.Gateway{
						{Namespace: "test", Name: "gateway-2"}: gw2,
					}
					expGraph.Routes[hrName] = expRouteHR2
					expGraph.Routes[hrName].ParentRefs[0].Attachment = &graph.ParentRefAttachmentStatus{
						AcceptedHostnames: map[string][]string{},
						FailedCondition:   conditions.NewTODO("Gateway is ignored"),
					}
					expGraph.Routes[hrName].ParentRefs[1].Attachment = &graph.ParentRefAttachmentStatus{
						AcceptedHostnames: map[string][]string{},
						FailedCondition:   conditions.NewTODO("Gateway is ignored"),
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the first Gateway is deleted", func() {
				It("returns updated graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.Gateway{},
						types.NamespacedName{Namespace: "test", Name: "gateway-1"},
					)

					// gateway 2 takes over;
					// route 1 has been replaced by route 2
					hr1Name := types.NamespacedName{Namespace: "test", Name: "hr-1"}
					hr2Name := types.NamespacedName{Namespace: "test", Name: "hr-2"}
					expGraph.Gateway.Source = gw2
					delete(expGraph.Gateway.Listeners["listener-80-1"].Routes, hr1Name)
					delete(expGraph.Gateway.Listeners["listener-443-1"].Routes, hr1Name)
					expGraph.Gateway.Listeners["listener-80-1"].Routes[hr2Name] = expRouteHR2
					expGraph.Gateway.Listeners["listener-443-1"].Routes[hr2Name] = expRouteHR2

					delete(expGraph.Routes, hr1Name)
					expGraph.Routes[hr2Name] = expRouteHR2

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the second HTTPRoute is deleted", func() {
				It("returns updated graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.HTTPRoute{},
						types.NamespacedName{Namespace: "test", Name: "hr-2"},
					)

					// gateway 2 still in charge;
					// no routes remain
					hr1Name := types.NamespacedName{Namespace: "test", Name: "hr-1"}
					expGraph.Gateway.Source = gw2
					delete(expGraph.Gateway.Listeners["listener-80-1"].Routes, hr1Name)
					delete(expGraph.Gateway.Listeners["listener-443-1"].Routes, hr1Name)
					expGraph.Routes = map[types.NamespacedName]*graph.Route{}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the GatewayClass is deleted", func() {
				It("returns updated graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.GatewayClass{},
						types.NamespacedName{Name: gcName},
					)

					expGraph.GatewayClass = nil
					expGraph.Gateway = &graph.Gateway{
						Source:     gw2,
						Conditions: conditions.NewGatewayInvalid("GatewayClass doesn't exist"),
					}
					expGraph.Routes = map[types.NamespacedName]*graph.Route{}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the second Gateway is deleted", func() {
				It("returns updated graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.Gateway{},
						types.NamespacedName{Namespace: "test", Name: "gateway-2"},
					)

					expGraph := &graph.Graph{}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the first HTTPRoute is deleted", func() {
				It("returns updated graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.HTTPRoute{},
						types.NamespacedName{Namespace: "test", Name: "hr-1"},
					)

					expGraph := &graph.Graph{}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
		})
		Describe("Process services and endpoints", Ordered, func() {
			var (
				hr1, hr2, hr3, hrInvalidBackendRef, hrMultipleRules                 *v1beta1.HTTPRoute
				hr1svc, sharedSvc, bazSvc1, bazSvc2, bazSvc3, invalidSvc, notRefSvc *apiv1.Service
				hr1slice1, hr1slice2, noRefSlice, missingSvcNameSlice               *discoveryV1.EndpointSlice
			)

			createSvc := func(name string) *apiv1.Service {
				return &apiv1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      name,
					},
				}
			}

			createEndpointSlice := func(name string, svcName string) *discoveryV1.EndpointSlice {
				return &discoveryV1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      name,
						Labels:    map[string]string{index.KubernetesServiceNameLabel: svcName},
					},
				}
			}

			BeforeAll(func() {
				testNamespace := v1beta1.Namespace("test")
				kindService := v1beta1.Kind("Service")
				kindInvalid := v1beta1.Kind("Invalid")

				// backend Refs
				fooRef := createBackendRef(&kindService, "foo-svc", &testNamespace)
				baz1NilNamespace := createBackendRef(&kindService, "baz-svc-v1", &testNamespace)
				barRef := createBackendRef(&kindService, "bar-svc", nil)
				baz2Ref := createBackendRef(&kindService, "baz-svc-v2", &testNamespace)
				baz3Ref := createBackendRef(&kindService, "baz-svc-v3", &testNamespace)
				invalidKindRef := createBackendRef(&kindInvalid, "bar-svc", &testNamespace)

				// httproutes
				hr1 = createRoute("hr1", "gw", "foo.example.com", fooRef)
				hr2 = createRoute("hr2", "gw", "bar.example.com", barRef)
				// hr3 shares the same backendRef as hr2
				hr3 = createRoute("hr3", "gw", "bar.2.example.com", barRef)
				hrInvalidBackendRef = createRoute("hr-invalid", "gw", "invalid.com", invalidKindRef)
				hrMultipleRules = createRouteWithMultipleRules(
					"hr-multiple-rules",
					"gw",
					"mutli.example.com",
					[]v1beta1.HTTPRouteRule{
						createHTTPRule("/baz-v1", baz1NilNamespace),
						createHTTPRule("/baz-v2", baz2Ref),
						createHTTPRule("/baz-v3", baz3Ref),
					},
				)

				// services
				hr1svc = createSvc("foo-svc")
				sharedSvc = createSvc("bar-svc")  // shared between hr2 and hr3
				invalidSvc = createSvc("invalid") // nsname matches invalid BackendRef
				notRefSvc = createSvc("not-ref")
				bazSvc1 = createSvc("baz-svc-v1")
				bazSvc2 = createSvc("baz-svc-v2")
				bazSvc3 = createSvc("baz-svc-v3")

				// endpoint slices
				hr1slice1 = createEndpointSlice("hr1-1", "foo-svc")
				hr1slice2 = createEndpointSlice("hr1-2", "foo-svc")
				noRefSlice = createEndpointSlice("no-ref", "no-ref")
				missingSvcNameSlice = createEndpointSlice("missing-svc-name", "")
			})

			testProcessChangedVal := func(expChanged bool) {
				changed, _ := processor.Process()
				Expect(changed).To(Equal(expChanged))
			}

			testUpsertTriggersChange := func(obj client.Object, expChanged bool) {
				processor.CaptureUpsertChange(obj)
				testProcessChangedVal(expChanged)
			}

			testDeleteTriggersChange := func(obj client.Object, nsname types.NamespacedName, expChanged bool) {
				processor.CaptureDeleteChange(obj, nsname)
				testProcessChangedVal(expChanged)
			}
			When("hr1 is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr1, true)
				})
			})
			When("a hr1 service is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr1svc, true)
				})
			})
			When("an hr1 endpoint slice is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr1slice1, true)
				})
			})
			When("an hr1 service is updated", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr1svc, true)
				})
			})
			When("another hr1 endpoint slice is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr1slice2, true)
				})
			})
			When("an endpoint slice with a missing svc name label is added", func() {
				It("should not trigger a change", func() {
					testUpsertTriggersChange(missingSvcNameSlice, false)
				})
			})
			When("an hr1 endpoint slice is deleted", func() {
				It("should trigger a change", func() {
					testDeleteTriggersChange(
						hr1slice1,
						types.NamespacedName{Namespace: hr1slice1.Namespace, Name: hr1slice1.Name},
						true,
					)
				})
			})
			When("the second hr1 endpoint slice is deleted", func() {
				It("should trigger a change", func() {
					testDeleteTriggersChange(
						hr1slice2,
						types.NamespacedName{Namespace: hr1slice2.Namespace, Name: hr1slice2.Name},
						true,
					)
				})
			})
			When("the second hr1 endpoint slice is recreated", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr1slice2, true)
				})
			})
			When("hr1 is deleted", func() {
				It("should trigger a change", func() {
					testDeleteTriggersChange(
						hr1,
						types.NamespacedName{Namespace: hr1.Namespace, Name: hr1.Name},
						true,
					)
				})
			})
			When("hr1 service is deleted", func() {
				It("should not trigger a change", func() {
					testDeleteTriggersChange(
						hr1svc,
						types.NamespacedName{Namespace: hr1svc.Namespace, Name: hr1svc.Name},
						false,
					)
				})
			})
			When("the second hr1 endpoint slice is deleted", func() {
				It("should not trigger a change", func() {
					testDeleteTriggersChange(
						hr1slice2,
						types.NamespacedName{Namespace: hr1slice2.Namespace, Name: hr1slice2.Name},
						false,
					)
				})
			})
			When("hr2 is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr2, true)
				})
			})
			When("a hr3, that shares a backend service with hr2, is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hr3, true)
				})
			})
			When("sharedSvc, a service referenced by both hr2 and hr3, is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(sharedSvc, true)
				})
			})
			When("hr2 is deleted", func() {
				It("should trigger a change", func() {
					testDeleteTriggersChange(
						hr2,
						types.NamespacedName{Namespace: hr2.Namespace, Name: hr2.Name},
						true,
					)
				})
			})
			When("sharedSvc is deleted", func() {
				It("should trigger a change", func() {
					testDeleteTriggersChange(
						sharedSvc,
						types.NamespacedName{Namespace: sharedSvc.Namespace, Name: sharedSvc.Name},
						true,
					)
				})
			})
			When("sharedSvc is recreated", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(sharedSvc, true)
				})
			})
			When("hr3 is deleted", func() {
				It("should trigger a change", func() {
					testDeleteTriggersChange(
						hr3,
						types.NamespacedName{Namespace: hr3.Namespace, Name: hr3.Name},
						true,
					)
				})
			})
			When("sharedSvc is deleted", func() {
				It("should not trigger a change", func() {
					testDeleteTriggersChange(
						sharedSvc,
						types.NamespacedName{Namespace: sharedSvc.Namespace, Name: sharedSvc.Name},
						false,
					)
				})
			})
			When("a service that is not referenced by any route is added", func() {
				It("should not trigger a change", func() {
					testUpsertTriggersChange(notRefSvc, false)
				})
			})
			When("a route with an invalid backend ref type is added", func() {
				It("should trigger a change", func() {
					testUpsertTriggersChange(hrInvalidBackendRef, true)
				})
			})
			When("a service with a namespace name that matches invalid backend ref is added", func() {
				It("should not trigger a change", func() {
					testUpsertTriggersChange(invalidSvc, false)
				})
			})
			When("an endpoint slice that is not owned by a referenced service is added", func() {
				It("should not trigger a change", func() {
					testUpsertTriggersChange(noRefSlice, false)
				})
			})
			When("an endpoint slice that is not owned by a referenced service is deleted", func() {
				It("should not trigger a change", func() {
					testDeleteTriggersChange(
						noRefSlice,
						types.NamespacedName{Namespace: noRefSlice.Namespace, Name: noRefSlice.Name},
						false,
					)
				})
			})
			Context("processing a route with multiple rules and three unique backend services", func() {
				When("route is added", func() {
					It("should trigger a change", func() {
						testUpsertTriggersChange(hrMultipleRules, true)
					})
				})
				When("first referenced service is added", func() {
					It("should trigger a change", func() {
						testUpsertTriggersChange(bazSvc1, true)
					})
				})
				When("second referenced service is added", func() {
					It("should trigger a change", func() {
						testUpsertTriggersChange(bazSvc2, true)
					})
				})
				When("first referenced service is deleted", func() {
					It("should trigger a change", func() {
						testDeleteTriggersChange(
							bazSvc1,
							types.NamespacedName{Namespace: bazSvc1.Namespace, Name: bazSvc1.Name},
							true,
						)
					})
				})
				When("first referenced service is recreated", func() {
					It("should trigger a change", func() {
						testUpsertTriggersChange(bazSvc1, true)
					})
				})
				When("third referenced service is added", func() {
					It("should trigger a change", func() {
						testUpsertTriggersChange(bazSvc3, true)
					})
				})
				When("third referenced service is updated", func() {
					It("should trigger a change", func() {
						testUpsertTriggersChange(bazSvc3, true)
					})
				})
				When("route is deleted", func() {
					It("should trigger a change", func() {
						testDeleteTriggersChange(
							hrMultipleRules,
							types.NamespacedName{
								Namespace: hrMultipleRules.Namespace,
								Name:      hrMultipleRules.Name,
							},
							true,
						)
					})
				})
				When("first referenced service is deleted", func() {
					It("should not trigger a change", func() {
						testDeleteTriggersChange(
							bazSvc1,
							types.NamespacedName{Namespace: bazSvc1.Namespace, Name: bazSvc1.Name},
							false,
						)
					})
				})
				When("second referenced service is deleted", func() {
					It("should not trigger a change", func() {
						testDeleteTriggersChange(
							bazSvc2,
							types.NamespacedName{Namespace: bazSvc2.Namespace, Name: bazSvc2.Name},
							false,
						)
					})
				})
				When("final referenced service is deleted", func() {
					It("should not trigger a change", func() {
						testDeleteTriggersChange(
							bazSvc3,
							types.NamespacedName{Namespace: bazSvc3.Namespace, Name: bazSvc3.Name},
							false,
						)
					})
				})
			})
		})
	})

	Describe("Ensuring non-changing changes don't override previously changing changes", func() {
		// Note: in these tests, we deliberately don't fully inspect the returned configuration and statuses
		// -- this is done in 'Normal cases of processing changes'

		var (
			processor                               *state.ChangeProcessorImpl
			fakeRelationshipCapturer                *relationshipfakes.FakeCapturer
			gcNsName, gwNsName, hrNsName, hr2NsName types.NamespacedName
			svcNsName, sliceNsName                  types.NamespacedName
			gc, gcUpdated                           *v1beta1.GatewayClass
			gw1, gw1Updated, gw2                    *v1beta1.Gateway
			hr1, hr1Updated, hr2                    *v1beta1.HTTPRoute
			svc                                     *apiv1.Service
			slice                                   *discoveryV1.EndpointSlice
		)

		BeforeEach(OncePerOrdered, func() {
			fakeSecretMemoryMgr := &secretsfakes.FakeSecretDiskMemoryManager{}
			fakeRelationshipCapturer = &relationshipfakes.FakeCapturer{}

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      "test.controller",
				GatewayClassName:     "my-class",
				SecretMemoryManager:  fakeSecretMemoryMgr,
				RelationshipCapturer: fakeRelationshipCapturer,
				Validators:           createAlwaysValidValidators(),
				Scheme:               createScheme(),
			})

			gcNsName = types.NamespacedName{Name: "my-class"}

			gc = &v1beta1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: gcNsName.Name,
				},
				Spec: v1beta1.GatewayClassSpec{
					ControllerName: "test.controller",
				},
			}

			gcUpdated = gc.DeepCopy()
			gcUpdated.Generation++

			gwNsName = types.NamespacedName{Namespace: "test", Name: "gw-1"}

			gw1 = &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: gwNsName.Namespace,
					Name:      gwNsName.Name,
				},
			}

			gw1Updated = gw1.DeepCopy()
			gw1Updated.Generation++

			gw2 = gw1.DeepCopy()
			gw2.Name = "gw-2"

			hrNsName = types.NamespacedName{Namespace: "test", Name: "hr-1"}

			hr1 = &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: hrNsName.Namespace,
					Name:      hrNsName.Name,
				},
			}

			hr1Updated = hr1.DeepCopy()
			hr1Updated.Generation++

			hr2NsName = types.NamespacedName{Namespace: "test", Name: "hr-2"}

			hr2 = hr1.DeepCopy()
			hr2.Name = hr2NsName.Name

			svcNsName = types.NamespacedName{Namespace: "test", Name: "svc"}

			svc = &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: svcNsName.Namespace,
					Name:      svcNsName.Name,
				},
			}

			sliceNsName = types.NamespacedName{Namespace: "test", Name: "slice"}

			slice = &discoveryV1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: sliceNsName.Namespace,
					Name:      sliceNsName.Name,
				},
			}
		})
		// Changing change - a change that makes processor.Process() report changed
		// Non-changing change - a change that doesn't do that
		// Related resource - a K8s resource that is related to a configured Gateway API resource
		// Unrelated resource - a K8s resource that is not related to a configured Gateway API resource

		// Note: in these tests, we deliberately don't fully inspect the returned configuration and statuses
		// -- this is done in 'Normal cases of processing changes'
		Describe("Multiple Gateway API resource changes", Ordered, func() {
			It("should report changed after multiple Upserts", func() {
				processor.CaptureUpsertChange(gc)
				processor.CaptureUpsertChange(gw1)
				processor.CaptureUpsertChange(hr1)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})
			It("should report not changed after multiple Upserts of the resource with same generation", func() {
				processor.CaptureUpsertChange(gc)
				processor.CaptureUpsertChange(gw1)
				processor.CaptureUpsertChange(hr1)

				changed, _ := processor.Process()
				Expect(changed).To(BeFalse())
			})
			When("a upsert of updated resources is followed by an upsert of the same generation", func() {
				It("should report changed", func() {
					// these are changing changes
					processor.CaptureUpsertChange(gcUpdated)
					processor.CaptureUpsertChange(gw1Updated)
					processor.CaptureUpsertChange(hr1Updated)

					// there are non-changing changes
					processor.CaptureUpsertChange(gcUpdated)
					processor.CaptureUpsertChange(gw1Updated)
					processor.CaptureUpsertChange(hr1Updated)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				})
			})
			It("should report changed after upserting new resources", func() {
				// we can't have a second GatewayClass, so we don't add it
				processor.CaptureUpsertChange(gw2)
				processor.CaptureUpsertChange(hr2)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})
			When("resources are deleted followed by upserts with the same generations", func() {
				It("should report changed", func() {
					// these are changing changes
					processor.CaptureDeleteChange(&v1beta1.GatewayClass{}, gcNsName)
					processor.CaptureDeleteChange(&v1beta1.Gateway{}, gwNsName)
					processor.CaptureDeleteChange(&v1beta1.HTTPRoute{}, hrNsName)

					// these are non-changing changes
					processor.CaptureUpsertChange(gw2)
					processor.CaptureUpsertChange(hr2)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				})
			})
			It("should report changed after deleting resources", func() {
				processor.CaptureDeleteChange(&v1beta1.HTTPRoute{}, hr2NsName)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})
		})
		Describe("Deleting non-existing Gateway API resource", func() {
			It("should not report changed after deleting non-existing", func() {
				processor.CaptureDeleteChange(&v1beta1.GatewayClass{}, gcNsName)
				processor.CaptureDeleteChange(&v1beta1.Gateway{}, gwNsName)
				processor.CaptureDeleteChange(&v1beta1.HTTPRoute{}, hrNsName)
				processor.CaptureDeleteChange(&v1beta1.HTTPRoute{}, hr2NsName)

				changed, _ := processor.Process()
				Expect(changed).To(BeFalse())
			})
		})
		Describe("Multiple Kubernetes API resource changes", Ordered, func() {
			It("should report changed after multiple Upserts of related resources", func() {
				fakeRelationshipCapturer.ExistsReturns(true)
				processor.CaptureUpsertChange(svc)
				processor.CaptureUpsertChange(slice)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})

			It("should report not changed after multiple Upserts of unrelated resources", func() {
				fakeRelationshipCapturer.ExistsReturns(false)
				processor.CaptureUpsertChange(svc)
				processor.CaptureUpsertChange(slice)

				changed, _ := processor.Process()
				Expect(changed).To(BeFalse())
			})
			When("upserts of related resources are followed by upserts of unrelated resources", func() {
				It("should report changed", func() {
					// these are changing changes
					fakeRelationshipCapturer.ExistsReturns(true)
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)

					// there are non-changing changes
					fakeRelationshipCapturer.ExistsReturns(false)
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				})
			})
			When("deletes of related resources are followed by upserts of unrelated resources", func() {
				It("should report changed", func() {
					// these are changing changes
					fakeRelationshipCapturer.ExistsReturns(true)
					processor.CaptureDeleteChange(&apiv1.Service{}, svcNsName)
					processor.CaptureDeleteChange(&discoveryV1.EndpointSlice{}, sliceNsName)

					// these are non-changing changes
					fakeRelationshipCapturer.ExistsReturns(false)
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				})
			})
		})
		Describe("Multiple Kubernetes API and Gateway API resource changes", Ordered, func() {
			It("should report changed after multiple Upserts of new and related resources", func() {
				// new Gateway API resources
				fakeRelationshipCapturer.ExistsReturns(false)
				processor.CaptureUpsertChange(gc)
				processor.CaptureUpsertChange(gw1)
				processor.CaptureUpsertChange(hr1)

				// related Kubernetes API resources
				fakeRelationshipCapturer.ExistsReturns(true)
				processor.CaptureUpsertChange(svc)
				processor.CaptureUpsertChange(slice)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})

			It("should report not changed after multiple Upserts of unrelated and unchanged resources", func() {
				// unchanged Gateway API resources
				fakeRelationshipCapturer.ExistsReturns(false)
				processor.CaptureUpsertChange(gc)
				processor.CaptureUpsertChange(gw1)
				processor.CaptureUpsertChange(hr1)

				// unrelated Kubernetes API resources
				fakeRelationshipCapturer.ExistsReturns(false)
				processor.CaptureUpsertChange(svc)
				processor.CaptureUpsertChange(slice)

				changed, _ := processor.Process()
				Expect(changed).To(BeFalse())
			})

			It("should report changed after upserting related resources followed by upserting unchanged resources",
				func() {
					// these are changing changes
					fakeRelationshipCapturer.ExistsReturns(true)
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)

					// these are non-changing changes
					fakeRelationshipCapturer.ExistsReturns(false)
					processor.CaptureUpsertChange(gc)
					processor.CaptureUpsertChange(gw1)
					processor.CaptureUpsertChange(hr1)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				},
			)

			It("should report changed after upserting changed resources followed by upserting unrelated resources",
				func() {
					// these are changing changes
					fakeRelationshipCapturer.ExistsReturns(false)
					processor.CaptureUpsertChange(gcUpdated)
					processor.CaptureUpsertChange(gw1Updated)
					processor.CaptureUpsertChange(hr1Updated)

					// these are non-changing changes
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				},
			)
			It(
				"should report changed after upserting related resources followed by upserting unchanged resources",
				func() {
					// these are changing changes
					fakeRelationshipCapturer.ExistsReturns(true)
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)

					// these are non-changing changes
					fakeRelationshipCapturer.ExistsReturns(false)
					processor.CaptureUpsertChange(gcUpdated)
					processor.CaptureUpsertChange(gw1Updated)
					processor.CaptureUpsertChange(hr1Updated)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				},
			)
		})
	})

	Describe("Webhook validation cases", Ordered, func() {
		var (
			processor           state.ChangeProcessor
			fakeEventRecorder   *record.FakeRecorder
			fakeSecretMemoryMgr *secretsfakes.FakeSecretDiskMemoryManager

			gc *v1beta1.GatewayClass

			gwNsName, hrNsName types.NamespacedName
			gw, gwInvalid      *v1beta1.Gateway
			hr, hrInvalid      *v1beta1.HTTPRoute
		)
		BeforeAll(func() {
			fakeSecretMemoryMgr = &secretsfakes.FakeSecretDiskMemoryManager{}
			fakeEventRecorder = record.NewFakeRecorder(2 /* number of buffered events */)

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      controllerName,
				GatewayClassName:     gcName,
				SecretMemoryManager:  fakeSecretMemoryMgr,
				RelationshipCapturer: relationship.NewCapturerImpl(),
				Logger:               zap.New(),
				Validators:           createAlwaysValidValidators(),
				EventRecorder:        fakeEventRecorder,
				Scheme:               createScheme(),
			})

			gc = &v1beta1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:       gcName,
					Generation: 1,
				},
				Spec: v1beta1.GatewayClassSpec{
					ControllerName: controllerName,
				},
			}

			gwNsName = types.NamespacedName{Namespace: "test", Name: "gateway"}
			hrNsName = types.NamespacedName{Namespace: "test", Name: "hr"}

			gw = &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: gwNsName.Namespace,
					Name:      gwNsName.Name,
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						{
							Name:     "listener-80-1",
							Hostname: helpers.GetPointer[v1beta1.Hostname]("foo.example.com"),
							Port:     80,
							Protocol: v1beta1.HTTPProtocolType,
						},
					},
				},
			}

			gwInvalid = gw.DeepCopy()
			// cannot have hostname for TCP protocol
			gwInvalid.Spec.Listeners[0].Protocol = v1beta1.TCPProtocolType

			hr = &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: hrNsName.Namespace,
					Name:      hrNsName.Name,
				},
				Spec: v1beta1.HTTPRouteSpec{
					CommonRouteSpec: v1beta1.CommonRouteSpec{
						ParentRefs: []v1beta1.ParentReference{
							{
								Namespace: (*v1beta1.Namespace)(&gw.Namespace),
								Name:      v1beta1.ObjectName(gw.Name),
								SectionName: (*v1beta1.SectionName)(
									helpers.GetPointer("listener-80-1"),
								),
							},
						},
					},
					Hostnames: []v1beta1.Hostname{
						"foo.example.com",
					},
					Rules: []v1beta1.HTTPRouteRule{
						{
							Matches: []v1beta1.HTTPRouteMatch{
								{
									Path: &v1beta1.HTTPPathMatch{
										Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
										Value: helpers.GetPointer("/"),
									},
								},
							},
						},
					},
				},
			}

			hrInvalid = hr.DeepCopy()
			hrInvalid.Spec.Rules[0].Matches[0].Path.Type = nil // cannot be nil
		})

		assertHREvent := func() {
			var e string
			EventuallyWithOffset(1, fakeEventRecorder.Events).Should(Receive(&e))
			ExpectWithOffset(1, e).To(ContainSubstring("Rejected"))
			ExpectWithOffset(1, e).To(ContainSubstring("spec.rules[0].matches[0].path.type"))
		}

		assertGwEvent := func() {
			var e string
			EventuallyWithOffset(1, fakeEventRecorder.Events).Should(Receive(&e))
			ExpectWithOffset(1, e).To(ContainSubstring("Rejected"))
			ExpectWithOffset(1, e).To(ContainSubstring("spec.listeners[0].hostname"))
		}

		It("should process GatewayClass", func() {
			processor.CaptureUpsertChange(gc)

			changed, graphCfg := processor.Process()
			Expect(changed).To(BeTrue())
			Expect(graphCfg.GatewayClass).ToNot(BeNil())
			Expect(fakeEventRecorder.Events).To(HaveLen(0))
		})

		When("resources are invalid", func() {
			It("should not process them", func() {
				processor.CaptureUpsertChange(gwInvalid)
				processor.CaptureUpsertChange(hrInvalid)

				changed, graphCfg := processor.Process()

				Expect(changed).To(BeFalse())
				Expect(graphCfg).To(BeNil())

				Expect(fakeEventRecorder.Events).To(HaveLen(2))
				assertGwEvent()
				assertHREvent()
			})
		})

		When("resources are valid", func() {
			It("should process them", func() {
				processor.CaptureUpsertChange(gw)
				processor.CaptureUpsertChange(hr)

				changed, graphCfg := processor.Process()

				Expect(changed).To(BeTrue())
				Expect(graphCfg).ToNot(BeNil())
				Expect(graphCfg.Gateway).ToNot(BeNil())
				Expect(graphCfg.Routes).To(HaveLen(1))

				Expect(fakeEventRecorder.Events).To(HaveLen(0))
			})
		})

		When("a new version of HTTPRoute is invalid", func() {
			It("it should delete the configuration for the old one and not process the new one", func() {
				processor.CaptureUpsertChange(hrInvalid)

				changed, graphCfg := processor.Process()

				Expect(changed).To(BeTrue())
				Expect(graphCfg.Routes).To(HaveLen(0))

				Expect(fakeEventRecorder.Events).To(HaveLen(1))
				assertHREvent()
			})
		})

		When("a new version of Gateway is invalid", func() {
			It("it should delete the configuration for the old one and not process the new one", func() {
				processor.CaptureUpsertChange(gwInvalid)

				changed, graphCfg := processor.Process()

				Expect(changed).To(BeTrue())
				Expect(graphCfg.Gateway).To(BeNil())

				Expect(fakeEventRecorder.Events).To(HaveLen(1))
				assertGwEvent()
			})
		})

		Describe("Webhook assumptions", func() {
			var processor state.ChangeProcessor

			BeforeEach(func() {
				fakeSecretMemoryMgr = &secretsfakes.FakeSecretDiskMemoryManager{}
				fakeEventRecorder = record.NewFakeRecorder(1 /* number of buffered events */)

				processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
					GatewayCtlrName:      controllerName,
					GatewayClassName:     gcName,
					SecretMemoryManager:  fakeSecretMemoryMgr,
					RelationshipCapturer: relationship.NewCapturerImpl(),
					Logger:               zap.New(),
					Validators:           createAlwaysValidValidators(),
					EventRecorder:        fakeEventRecorder,
					Scheme:               createScheme(),
				})
			})

			createInvalidHTTPRoute := func(invalidator func(hr *v1beta1.HTTPRoute)) *v1beta1.HTTPRoute {
				hr := createRoute(
					"hr",
					"gateway",
					"foo.example.com",
					createBackendRef(
						helpers.GetPointer[v1beta1.Kind]("Service"),
						"test",
						helpers.GetPointer[v1beta1.Namespace]("namespace"),
					),
				)
				invalidator(hr)
				return hr
			}

			createInvalidGateway := func(invalidator func(gw *v1beta1.Gateway)) *v1beta1.Gateway {
				gw := createGateway("gateway")
				invalidator(gw)
				return gw
			}

			assertRejectedEvent := func() {
				EventuallyWithOffset(1, fakeEventRecorder.Events).Should(Receive(ContainSubstring("Rejected")))
			}

			DescribeTable("Invalid HTTPRoutes",
				func(hr *v1beta1.HTTPRoute) {
					processor.CaptureUpsertChange(hr)

					changed, graphCfg := processor.Process()

					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())

					assertRejectedEvent()
				},
				Entry(
					"duplicate parentRefs",
					createInvalidHTTPRoute(func(hr *v1beta1.HTTPRoute) {
						hr.Spec.ParentRefs = append(hr.Spec.ParentRefs, hr.Spec.ParentRefs[len(hr.Spec.ParentRefs)-1])
					}),
				),
				Entry(
					"nil path.Type",
					createInvalidHTTPRoute(func(hr *v1beta1.HTTPRoute) {
						hr.Spec.Rules[0].Matches[0].Path.Type = nil
					}),
				),
				Entry("nil path.Value",
					createInvalidHTTPRoute(func(hr *v1beta1.HTTPRoute) {
						hr.Spec.Rules[0].Matches[0].Path.Value = nil
					}),
				),
				Entry(
					"nil request.Redirect",
					createInvalidHTTPRoute(func(hr *v1beta1.HTTPRoute) {
						hr.Spec.Rules[0].Filters = append(hr.Spec.Rules[0].Filters, v1beta1.HTTPRouteFilter{
							Type:            v1beta1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: nil,
						})
					}),
				),
				Entry("nil port in BackendRef",
					createInvalidHTTPRoute(func(hr *v1beta1.HTTPRoute) {
						hr.Spec.Rules[0].BackendRefs[0].Port = nil
					}),
				),
			)

			DescribeTable("Invalid Gateway resources",
				func(gw *v1beta1.Gateway) {
					processor.CaptureUpsertChange(gw)

					changed, graphCfg := processor.Process()

					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())

					assertRejectedEvent()
				},
				Entry("tls in HTTP listener",
					createInvalidGateway(func(gw *v1beta1.Gateway) {
						gw.Spec.Listeners[0].TLS = &v1beta1.GatewayTLSConfig{}
					}),
				),
				Entry("tls is nil in HTTPS listener",
					createInvalidGateway(func(gw *v1beta1.Gateway) {
						gw.Spec.Listeners[0].Protocol = v1beta1.HTTPSProtocolType
						gw.Spec.Listeners[0].TLS = nil
					}),
				),
				Entry("zero certificateRefs in HTTPS listener",
					createInvalidGateway(func(gw *v1beta1.Gateway) {
						gw.Spec.Listeners[0].Protocol = v1beta1.HTTPSProtocolType
						gw.Spec.Listeners[0].TLS = &v1beta1.GatewayTLSConfig{
							Mode:            helpers.GetPointer(v1beta1.TLSModeTerminate),
							CertificateRefs: nil,
						}
					}),
				),
			)
		})
	})

	Describe("Edge cases with panic", func() {
		var (
			processor                state.ChangeProcessor
			fakeSecretMemoryMgr      *secretsfakes.FakeSecretDiskMemoryManager
			fakeRelationshipCapturer *relationshipfakes.FakeCapturer
		)

		BeforeEach(func() {
			fakeSecretMemoryMgr = &secretsfakes.FakeSecretDiskMemoryManager{}
			fakeRelationshipCapturer = &relationshipfakes.FakeCapturer{}

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      "test.controller",
				GatewayClassName:     "my-class",
				SecretMemoryManager:  fakeSecretMemoryMgr,
				RelationshipCapturer: fakeRelationshipCapturer,
				Validators:           createAlwaysValidValidators(),
				Scheme:               createScheme(),
			})
		})

		DescribeTable("CaptureUpsertChange must panic",
			func(obj client.Object) {
				process := func() {
					processor.CaptureUpsertChange(obj)
				}
				Expect(process).Should(Panic())
			},
			Entry(
				"an unsupported resource",
				&v1alpha2.TCPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "tcp"}},
			),
			Entry(
				"nil resource",
				nil,
			),
		)

		DescribeTable(
			"CaptureDeleteChange must panic",
			func(resourceType client.Object, nsname types.NamespacedName) {
				process := func() {
					processor.CaptureDeleteChange(resourceType, nsname)
				}
				Expect(process).Should(Panic())
			},
			Entry(
				"an unsupported resource",
				&v1alpha2.TCPRoute{},
				types.NamespacedName{Namespace: "test", Name: "tcp"},
			),
			Entry(
				"nil resource type",
				nil,
				types.NamespacedName{Namespace: "test", Name: "resource"},
			),
		)
	})
})
