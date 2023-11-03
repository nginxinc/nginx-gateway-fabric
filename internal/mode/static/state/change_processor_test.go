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

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/index"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/relationship"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/relationship/relationshipfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation/validationfakes"
)

const (
	controllerName = "my.controller"
	gcName         = "test-class"
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

func createGatewayWithTLSListener(name string, tlsSecret *apiv1.Secret) *v1beta1.Gateway {
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
					Name:      v1beta1.ObjectName(tlsSecret.Name),
					Namespace: (*v1beta1.Namespace)(&tlsSecret.Namespace),
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
	utilruntime.Must(ngfAPI.AddToScheme(scheme))

	return scheme
}

var (
	cert = []byte(`-----BEGIN CERTIFICATE-----
MIIDLjCCAhYCCQDAOF9tLsaXWjANBgkqhkiG9w0BAQsFADBaMQswCQYDVQQGEwJV
UzELMAkGA1UECAwCQ0ExITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0
ZDEbMBkGA1UEAwwSY2FmZS5leGFtcGxlLmNvbSAgMB4XDTE4MDkxMjE2MTUzNVoX
DTIzMDkxMTE2MTUzNVowWDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxGTAXBgNVBAMMEGNhZmUuZXhh
bXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCp6Kn7sy81
p0juJ/cyk+vCAmlsfjtFM2muZNK0KtecqG2fjWQb55xQ1YFA2XOSwHAYvSdwI2jZ
ruW8qXXCL2rb4CZCFxwpVECrcxdjm3teViRXVsYImmJHPPSyQgpiobs9x7DlLc6I
BA0ZjUOyl0PqG9SJexMV73WIIa5rDVSF2r4kSkbAj4Dcj7LXeFlVXH2I5XwXCptC
n67JCg42f+k8wgzcRVp8XZkZWZVjwq9RUKDXmFB2YyN1XEWdZ0ewRuKYUJlsm692
skOrKQj0vkoPn41EE/+TaVEpqLTRoUY3rzg7DkdzfdBizFO2dsPNFx2CW0jXkNLv
Ko25CZrOhXAHAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKHFCcyOjZvoHswUBMdL
RdHIb383pWFynZq/LuUovsVA58B0Cg7BEfy5vWVVrq5RIkv4lZ81N29x21d1JH6r
jSnQx+DXCO/TJEV5lSCUpIGzEUYaUPgRyjsM/NUdCJ8uHVhZJ+S6FA+CnOD9rn2i
ZBePCI5rHwEXwnnl8ywij3vvQ5zHIuyBglWr/Qyui9fjPpwWUvUm4nv5SMG9zCV7
PpuwvuatqjO1208BjfE/cZHIg8Hw9mvW9x9C+IQMIMDE7b/g6OcK7LGTLwlFxvA8
7WjEequnayIphMhKRXVf1N349eN98Ez38fOTHTPbdJjFA/PcC+Gyme+iGt5OQdFh
yRE=
-----END CERTIFICATE-----`)
	key = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAqeip+7MvNadI7if3MpPrwgJpbH47RTNprmTStCrXnKhtn41k
G+ecUNWBQNlzksBwGL0ncCNo2a7lvKl1wi9q2+AmQhccKVRAq3MXY5t7XlYkV1bG
CJpiRzz0skIKYqG7Pcew5S3OiAQNGY1DspdD6hvUiXsTFe91iCGuaw1Uhdq+JEpG
wI+A3I+y13hZVVx9iOV8FwqbQp+uyQoONn/pPMIM3EVafF2ZGVmVY8KvUVCg15hQ
dmMjdVxFnWdHsEbimFCZbJuvdrJDqykI9L5KD5+NRBP/k2lRKai00aFGN684Ow5H
c33QYsxTtnbDzRcdgltI15DS7yqNuQmazoVwBwIDAQABAoIBAQCPSdSYnQtSPyql
FfVFpTOsoOYRhf8sI+ibFxIOuRauWehhJxdm5RORpAzmCLyL5VhjtJme223gLrw2
N99EjUKb/VOmZuDsBc6oCF6QNR58dz8cnORTewcotsJR1pn1hhlnR5HqJJBJask1
ZEnUQfcXZrL94lo9JH3E+Uqjo1FFs8xxE8woPBqjZsV7pRUZgC3LhxnwLSExyFo4
cxb9SOG5OmAJozStFoQ2GJOes8rJ5qfdvytgg9xbLaQL/x0kpQ62BoFMBDdqOePW
KfP5zZ6/07/vpj48yA1Q32PzobubsBLd3Kcn32jfm1E7prtWl+JeOFiOznBQFJbN
4qPVRz5hAoGBANtWyxhNCSLu4P+XgKyckljJ6F5668fNj5CzgFRqJ09zn0TlsNro
FTLZcxDqnR3HPYM42JERh2J/qDFZynRQo3cg3oeivUdBVGY8+FI1W0qdub/L9+yu
edOZTQ5XmGGp6r6jexymcJim/OsB3ZnYOpOrlD7SPmBvzNLk4MF6gxbXAoGBAMZO
0p6HbBmcP0tjFXfcKE77ImLm0sAG4uHoUx0ePj/2qrnTnOBBNE4MvgDuTJzy+caU
k8RqmdHCbHzTe6fzYq/9it8sZ77KVN1qkbIcuc+RTxA9nNh1TjsRne74Z0j1FCLk
hHcqH0ri7PYSKHTE8FvFCxZYdbuB84CmZihvxbpRAoGAIbjqaMYPTYuklCda5S79
YSFJ1JzZe1Kja//tDw1zFcgVCKa31jAwciz0f/lSRq3HS1GGGmezhPVTiqLfeZqc
R0iKbhgbOcVVkJJ3K0yAyKwPTumxKHZ6zImZS0c0am+RY9YGq5T7YrzpzcfvpiOU
ffe3RyFT7cfCmfoOhDCtzukCgYB30oLC1RLFOrqn43vCS51zc5zoY44uBzspwwYN
TwvP/ExWMf3VJrDjBCH+T/6sysePbJEImlzM+IwytFpANfiIXEt/48Xf60Nx8gWM
uHyxZZx/NKtDw0V8vX1POnq2A5eiKa+8jRARYKJLYNdfDuwolxvG6bZhkPi/4EtT
3Y18sQKBgHtKbk+7lNJVeswXE5cUG6EDUsDe/2Ua7fXp7FcjqBEoap1LSw+6TXp0
ZgrmKE8ARzM47+EJHUviiq/nupE15g0kJW3syhpU9zZLO7ltB0KIkO9ZRcmUjo8Q
cpLlHMAqbLJ8WYGJCkhiWxyal6hYTyWY4cVkC0xtTl/hUE9IeNKo
-----END RSA PRIVATE KEY-----`)
)

var _ = Describe("ChangeProcessor", func() {
	// graph outputs are large, so allow gomega to print everything on test failure
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
			processor state.ChangeProcessor
		)

		BeforeEach(OncePerOrdered, func() {
			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      controllerName,
				GatewayClassName:     gcName,
				RelationshipCapturer: relationship.NewCapturerImpl(gcName),
				Logger:               zap.New(),
				Validators:           createAlwaysValidValidators(),
				Scheme:               createScheme(),
			})
		})

		Describe("Process gateway resources", Ordered, func() {
			var (
				gcUpdated                        *v1beta1.GatewayClass
				diffNsTLSSecret, sameNsTLSSecret *apiv1.Secret
				hr1, hr1Updated, hr2             *v1beta1.HTTPRoute
				gw1, gw1Updated, gw2             *v1beta1.Gateway
				refGrant1, refGrant2             *v1beta1.ReferenceGrant
				expGraph                         *graph.Graph
				expRouteHR1, expRouteHR2         *graph.Route
				hr1Name, hr2Name                 types.NamespacedName
			)
			BeforeAll(func() {
				gcUpdated = gc.DeepCopy()
				gcUpdated.Generation++

				crossNsBackendRef := v1beta1.HTTPBackendRef{
					BackendRef: v1beta1.BackendRef{
						BackendObjectReference: v1beta1.BackendObjectReference{
							Kind:      helpers.GetPointer[v1beta1.Kind]("Service"),
							Name:      "service",
							Namespace: helpers.GetPointer[v1beta1.Namespace]("service-ns"),
							Port:      helpers.GetPointer[v1beta1.PortNumber](80),
						},
					},
				}

				hr1 = createRoute("hr-1", "gateway-1", "foo.example.com", crossNsBackendRef)
				hr1Name = types.NamespacedName{Namespace: hr1.Namespace, Name: hr1.Name}

				hr1Updated = hr1.DeepCopy()
				hr1Updated.Generation++

				hr2 = createRoute("hr-2", "gateway-2", "bar.example.com")
				hr2Name = types.NamespacedName{Namespace: "test", Name: "hr-2"}

				refGrant1 = &v1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "cert-ns",
						Name:      "ref-grant",
					},
					Spec: v1beta1.ReferenceGrantSpec{
						From: []v1beta1.ReferenceGrantFrom{
							{
								Group:     v1beta1.GroupName,
								Kind:      "Gateway",
								Namespace: "test",
							},
						},
						To: []v1beta1.ReferenceGrantTo{
							{
								Kind: "Secret",
							},
						},
					},
				}

				refGrant2 = &v1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "service-ns",
						Name:      "ref-grant",
					},
					Spec: v1beta1.ReferenceGrantSpec{
						From: []v1beta1.ReferenceGrantFrom{
							{
								Group:     v1beta1.GroupName,
								Kind:      "HTTPRoute",
								Namespace: "test",
							},
						},
						To: []v1beta1.ReferenceGrantTo{
							{
								Kind: "Service",
							},
						},
					},
				}

				sameNsTLSSecret = &apiv1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tls-secret",
						Namespace: "test",
					},
					Type: apiv1.SecretTypeTLS,
					Data: map[string][]byte{
						apiv1.TLSCertKey:       cert,
						apiv1.TLSPrivateKeyKey: key,
					},
				}

				diffNsTLSSecret = &apiv1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "different-ns-tls-secret",
						Namespace: "cert-ns",
					},
					Type: apiv1.SecretTypeTLS,
					Data: map[string][]byte{
						apiv1.TLSCertKey:       cert,
						apiv1.TLSPrivateKeyKey: key,
					},
				}

				gw1 = createGatewayWithTLSListener("gateway-1", diffNsTLSSecret) // cert in diff namespace than gw

				gw1Updated = gw1.DeepCopy()
				gw1Updated.Generation++

				gw2 = createGatewayWithTLSListener("gateway-2", sameNsTLSSecret)
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
					Rules: []graph.Rule{
						{
							BackendRefs: []graph.BackendRef{
								{
									Weight: 1,
								},
							},
							ValidMatches: true,
							ValidFilters: true,
						},
					},
					Valid: true,
					Conditions: []conditions.Condition{
						staticConds.NewRouteBackendRefRefBackendNotFound(
							"spec.rules[0].backendRefs[0].name: Not found: \"service\"",
						),
					},
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
								SupportedKinds: []v1beta1.RouteGroupKind{{Kind: "HTTPRoute"}},
							},
							"listener-443-1": {
								Source: gw1.Spec.Listeners[1],
								Valid:  true,
								Routes: map[types.NamespacedName]*graph.Route{
									{Namespace: "test", Name: "hr-1"}: expRouteHR1,
								},
								ResolvedSecret: helpers.GetPointer(client.ObjectKeyFromObject(diffNsTLSSecret)),
								SupportedKinds: []v1beta1.RouteGroupKind{{Kind: "HTTPRoute"}},
							},
						},
						Valid: true,
					},
					IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{},
					Routes: map[types.NamespacedName]*graph.Route{
						{Namespace: "test", Name: "hr-1"}: expRouteHR1,
					},
					ReferencedSecrets: map[types.NamespacedName]*graph.Secret{},
				}
			})

			When("no upsert has occurred", func() {
				It("returns nil graph", func() {
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
							Expect(helpers.Diff(&graph.Graph{}, graphCfg)).To(BeEmpty())
						})
					})
					When("the different namespace TLS Secret is upserted", func() {
						It("returns nil graph", func() {
							processor.CaptureUpsertChange(diffNsTLSSecret)

							changed, graphCfg := processor.Process()
							Expect(changed).To(BeFalse())
							Expect(graphCfg).To(BeNil())
						})
					})
					When("the first Gateway is upserted", func() {
						It("returns populated graph", func() {
							processor.CaptureUpsertChange(gw1)

							expGraph.GatewayClass = nil

							expGraph.Gateway.Conditions = staticConds.NewGatewayInvalid("GatewayClass doesn't exist")
							expGraph.Gateway.Valid = false
							expGraph.Gateway.Listeners = nil

							// no ref grant exists yet for hr1
							expGraph.Routes[hr1Name].Conditions = []conditions.Condition{
								staticConds.NewRouteBackendRefRefNotPermitted(
									"Backend ref to Service service-ns/service not permitted by any ReferenceGrant",
								),
							}
							expGraph.Routes[hr1Name].ParentRefs[0].Attachment = &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{},
								FailedCondition:   staticConds.NewRouteInvalidGateway(),
							}
							expGraph.Routes[hr1Name].ParentRefs[1].Attachment = &graph.ParentRefAttachmentStatus{
								AcceptedHostnames: map[string][]string{},
								FailedCondition:   staticConds.NewRouteInvalidGateway(),
							}

							expGraph.ReferencedSecrets = nil

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

					// no ref grant exists yet for gw1
					expGraph.Gateway.Listeners["listener-443-1"] = &graph.Listener{
						Source: gw1.Spec.Listeners[1],
						Valid:  false,
						Routes: map[types.NamespacedName]*graph.Route{},
						Conditions: staticConds.NewListenerRefNotPermitted(
							"Certificate ref to secret cert-ns/different-ns-tls-secret not permitted by any ReferenceGrant",
						),
						SupportedKinds: []v1beta1.RouteGroupKind{{Kind: "HTTPRoute"}},
					}

					expAttachment := &graph.ParentRefAttachmentStatus{
						AcceptedHostnames: map[string][]string{},
						FailedCondition:   staticConds.NewRouteInvalidListener(),
						Attached:          false,
					}

					expGraph.Gateway.Listeners["listener-80-1"].Routes[hr1Name].ParentRefs[1].Attachment = expAttachment

					// no ref grant exists yet for hr1
					expGraph.Routes[hr1Name].ParentRefs[1].Attachment = expAttachment
					expGraph.Routes[hr1Name].Conditions = []conditions.Condition{
						staticConds.NewRouteBackendRefRefNotPermitted(
							"Backend ref to Service service-ns/service not permitted by any ReferenceGrant",
						),
					}

					expGraph.ReferencedSecrets = nil

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the ReferenceGrant allowing the Gateway to reference its Secret is upserted", func() {
				It("returns updated graph", func() {
					processor.CaptureUpsertChange(refGrant1)

					// no ref grant exists yet for hr1
					expGraph.Routes[hr1Name].Conditions = []conditions.Condition{
						staticConds.NewRouteBackendRefRefNotPermitted(
							"Backend ref to Service service-ns/service not permitted by any ReferenceGrant",
						),
					}
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})

			When("the ReferenceGrant allowing the hr1 to reference the Service in different ns is upserted", func() {
				It("returns updated graph", func() {
					processor.CaptureUpsertChange(refGrant2)

					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the first HTTPRoute without a generation changed is processed", func() {
				It("returns nil graph", func() {
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

					expGraph.Gateway.Listeners["listener-443-1"].Routes[hr1Name].Source.Generation = hr1Updated.Generation
					expGraph.Gateway.Listeners["listener-80-1"].Routes[hr1Name].Source.Generation = hr1Updated.Generation
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				},
				)
			})
			When("the first Gateway update without generation changed is processed", func() {
				It("returns nil graph", func() {
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
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the GatewayClass update without generation change is processed", func() {
				It("returns nil graph", func() {
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
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the different namespace TLS secret is upserted again", func() {
				It("returns populated graph", func() {
					processor.CaptureUpsertChange(diffNsTLSSecret)

					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("no changes are captured", func() {
				It("returns nil graph", func() {
					changed, graphCfg := processor.Process()

					Expect(changed).To(BeFalse())
					Expect(graphCfg).To(BeNil())
				})
			})
			When("the same namespace TLS Secret is upserted", func() {
				It("returns nil graph", func() {
					processor.CaptureUpsertChange(sameNsTLSSecret)

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
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
					}

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the second HTTPRoute is upserted", func() {
				It("returns populated graph", func() {
					processor.CaptureUpsertChange(hr2)

					expGraph.IgnoredGateways = map[types.NamespacedName]*v1beta1.Gateway{
						{Namespace: "test", Name: "gateway-2"}: gw2,
					}
					expGraph.Routes[hr2Name] = expRouteHR2
					expGraph.Routes[hr2Name].ParentRefs[0].Attachment = &graph.ParentRefAttachmentStatus{
						AcceptedHostnames: map[string][]string{},
						FailedCondition:   staticConds.NewTODO("Gateway is ignored"),
					}
					expGraph.Routes[hr2Name].ParentRefs[1].Attachment = &graph.ParentRefAttachmentStatus{
						AcceptedHostnames: map[string][]string{},
						FailedCondition:   staticConds.NewTODO("Gateway is ignored"),
					}
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(diffNsTLSSecret)] = &graph.Secret{
						Source: diffNsTLSSecret,
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
					expGraph.Gateway.Source = gw2
					expGraph.Gateway.Listeners["listener-80-1"].Source = gw2.Spec.Listeners[0]
					expGraph.Gateway.Listeners["listener-443-1"].Source = gw2.Spec.Listeners[1]
					delete(expGraph.Gateway.Listeners["listener-80-1"].Routes, hr1Name)
					delete(expGraph.Gateway.Listeners["listener-443-1"].Routes, hr1Name)
					expGraph.Gateway.Listeners["listener-80-1"].Routes[hr2Name] = expRouteHR2
					expGraph.Gateway.Listeners["listener-443-1"].Routes[hr2Name] = expRouteHR2
					delete(expGraph.Routes, hr1Name)
					expGraph.Routes[hr2Name] = expRouteHR2
					sameNsTLSSecretRef := helpers.GetPointer(client.ObjectKeyFromObject(sameNsTLSSecret))
					expGraph.Gateway.Listeners["listener-443-1"].ResolvedSecret = sameNsTLSSecretRef
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(sameNsTLSSecret)] = &graph.Secret{
						Source: sameNsTLSSecret,
					}

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
					expGraph.Gateway.Source = gw2
					expGraph.Gateway.Listeners["listener-80-1"].Source = gw2.Spec.Listeners[0]
					expGraph.Gateway.Listeners["listener-443-1"].Source = gw2.Spec.Listeners[1]
					delete(expGraph.Gateway.Listeners["listener-80-1"].Routes, hr1Name)
					delete(expGraph.Gateway.Listeners["listener-443-1"].Routes, hr1Name)
					expGraph.Routes = map[types.NamespacedName]*graph.Route{}
					sameNsTLSSecretRef := helpers.GetPointer(client.ObjectKeyFromObject(sameNsTLSSecret))
					expGraph.Gateway.Listeners["listener-443-1"].ResolvedSecret = sameNsTLSSecretRef
					expGraph.ReferencedSecrets[client.ObjectKeyFromObject(sameNsTLSSecret)] = &graph.Secret{
						Source: sameNsTLSSecret,
					}

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
						Conditions: staticConds.NewGatewayInvalid("GatewayClass doesn't exist"),
					}
					expGraph.Routes = map[types.NamespacedName]*graph.Route{}
					expGraph.ReferencedSecrets = nil

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(expGraph, graphCfg)).To(BeEmpty())
				})
			})
			When("the second Gateway is deleted", func() {
				It("returns empty graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.Gateway{},
						types.NamespacedName{Namespace: "test", Name: "gateway-2"},
					)

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(&graph.Graph{}, graphCfg)).To(BeEmpty())
				})
			})
			When("the first HTTPRoute is deleted", func() {
				It("returns empty graph", func() {
					processor.CaptureDeleteChange(
						&v1beta1.HTTPRoute{},
						types.NamespacedName{Namespace: "test", Name: "hr-1"},
					)

					changed, graphCfg := processor.Process()
					Expect(changed).To(BeTrue())
					Expect(helpers.Diff(&graph.Graph{}, graphCfg)).To(BeEmpty())
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
		Describe("namespace changes", func() {
			When("namespace is linked via label selectors", func() {
				It("triggers an update when labels are removed", func() {
					ns := &apiv1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ns",
							Labels: map[string]string{
								"app": "allowed",
							},
						},
					}
					gw := &v1beta1.Gateway{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gw",
						},
						Spec: v1beta1.GatewaySpec{
							Listeners: []v1beta1.Listener{
								{
									AllowedRoutes: &v1beta1.AllowedRoutes{
										Namespaces: &v1beta1.RouteNamespaces{
											From: helpers.GetPointer(v1beta1.NamespacesFromSelector),
											Selector: &metav1.LabelSelector{
												MatchLabels: map[string]string{
													"app": "allowed",
												},
											},
										},
									},
								},
							},
						},
					}

					processor.CaptureUpsertChange(gw)
					processor.CaptureUpsertChange(ns)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())

					newNS := ns.DeepCopy()
					newNS.Labels = nil
					processor.CaptureUpsertChange(newNS)

					changed, _ = processor.Process()
					Expect(changed).To(BeTrue())
				})
			})
		})
		Describe("NginxProxy resource changes", Ordered, func() {
			paramGC := gc.DeepCopy()
			paramGC.Spec.ParametersRef = &v1beta1.ParametersReference{
				Group:     ngfAPI.GroupName,
				Kind:      v1beta1.Kind("NginxProxy"),
				Name:      "np",
				Namespace: helpers.GetPointer(v1beta1.Namespace("test")),
			}

			np := &ngfAPI.NginxProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "np",
					Namespace: "test",
				},
			}
			It("handles upserts for an NginxProxy", func() {
				processor.CaptureUpsertChange(np)
				processor.CaptureUpsertChange(paramGC)

				changed, graph := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(graph.NginxProxy).To(Equal(np))

				// TODO(sberman): Once some fields actually exist
				// for this resource, verify that an update occurs.
			})
			It("handles deletes for an NginxProxy", func() {
				processor.CaptureDeleteChange(np, client.ObjectKeyFromObject(np))

				changed, graph := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(graph.NginxProxy).To(BeNil())
			})
		})
	})

	Describe("Ensuring non-changing changes don't override previously changing changes", func() {
		// Note: in these tests, we deliberately don't fully inspect the returned configuration and statuses
		// -- this is done in 'Normal cases of processing changes'

		var (
			processor                                         *state.ChangeProcessorImpl
			fakeRelationshipCapturer                          *relationshipfakes.FakeCapturer
			gcNsName, gwNsName, hrNsName, hr2NsName, rgNsName types.NamespacedName
			svcNsName, sliceNsName, secretNsName              types.NamespacedName
			gc, gcUpdated                                     *v1beta1.GatewayClass
			gw1, gw1Updated, gw2                              *v1beta1.Gateway
			hr1, hr1Updated, hr2                              *v1beta1.HTTPRoute
			rg1, rg1Updated, rg2                              *v1beta1.ReferenceGrant
			svc                                               *apiv1.Service
			slice                                             *discoveryV1.EndpointSlice
			secret                                            *apiv1.Secret
		)

		BeforeEach(OncePerOrdered, func() {
			fakeRelationshipCapturer = &relationshipfakes.FakeCapturer{}

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      "test.controller",
				GatewayClassName:     "my-class",
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

			rgNsName = types.NamespacedName{Namespace: "test", Name: "rg-1"}

			rg1 = &v1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      rgNsName.Name,
					Namespace: rgNsName.Namespace,
				},
			}

			rg1Updated = rg1.DeepCopy()
			rg1Updated.Generation++

			rg2 = rg1.DeepCopy()
			rg2.Name = "rg-2"

			secretNsName = types.NamespacedName{Namespace: "test", Name: "test-secret"}

			secret = &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: secretNsName.Namespace,
					Name:      secretNsName.Name,
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
				processor.CaptureUpsertChange(rg1)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})
			It("should report not changed after multiple Upserts of the resource with same generation", func() {
				processor.CaptureUpsertChange(gc)
				processor.CaptureUpsertChange(gw1)
				processor.CaptureUpsertChange(hr1)
				processor.CaptureUpsertChange(rg1)

				changed, _ := processor.Process()
				Expect(changed).To(BeFalse())
			})
			When("a upsert of updated resources is followed by an upsert of the same generation", func() {
				It("should report changed", func() {
					// these are changing changes
					processor.CaptureUpsertChange(gcUpdated)
					processor.CaptureUpsertChange(gw1Updated)
					processor.CaptureUpsertChange(hr1Updated)
					processor.CaptureUpsertChange(rg1Updated)

					// there are non-changing changes
					processor.CaptureUpsertChange(gcUpdated)
					processor.CaptureUpsertChange(gw1Updated)
					processor.CaptureUpsertChange(hr1Updated)
					processor.CaptureUpsertChange(rg1Updated)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				})
			})
			It("should report changed after upserting new resources", func() {
				// we can't have a second GatewayClass, so we don't add it
				processor.CaptureUpsertChange(gw2)
				processor.CaptureUpsertChange(hr2)
				processor.CaptureUpsertChange(rg2)

				changed, _ := processor.Process()
				Expect(changed).To(BeTrue())
			})
			When("resources are deleted followed by upserts with the same generations", func() {
				It("should report changed", func() {
					// these are changing changes
					processor.CaptureDeleteChange(&v1beta1.GatewayClass{}, gcNsName)
					processor.CaptureDeleteChange(&v1beta1.Gateway{}, gwNsName)
					processor.CaptureDeleteChange(&v1beta1.HTTPRoute{}, hrNsName)
					processor.CaptureDeleteChange(&v1beta1.ReferenceGrant{}, rgNsName)

					// these are non-changing changes
					processor.CaptureUpsertChange(gw2)
					processor.CaptureUpsertChange(hr2)
					processor.CaptureUpsertChange(rg2)

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
				processor.CaptureDeleteChange(&v1beta1.ReferenceGrant{}, rgNsName)

				changed, _ := processor.Process()
				Expect(changed).To(BeFalse())
			})
		})
		Describe("Multiple Kubernetes API resource changes", Ordered, func() {
			// Note: because secret resource is not used by the real relationship.Capturer, it is not used
			// in the same way as service and endpoint slice in the tests below.
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

				processor.CaptureUpsertChange(secret)

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
					processor.CaptureUpsertChange(secret)

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
					processor.CaptureUpsertChange(secret)

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
				processor.CaptureUpsertChange(rg1)

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
				processor.CaptureUpsertChange(rg1)

				// unrelated Kubernetes API resources
				fakeRelationshipCapturer.ExistsReturns(false)
				processor.CaptureUpsertChange(svc)
				processor.CaptureUpsertChange(slice)
				processor.CaptureUpsertChange(secret)

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
					processor.CaptureUpsertChange(rg1)
					processor.CaptureUpsertChange(secret)

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
					processor.CaptureUpsertChange(rg1Updated)

					// these are non-changing changes
					processor.CaptureUpsertChange(svc)
					processor.CaptureUpsertChange(slice)
					processor.CaptureUpsertChange(secret)

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
					processor.CaptureUpsertChange(rg1Updated)
					processor.CaptureUpsertChange(secret)

					changed, _ := processor.Process()
					Expect(changed).To(BeTrue())
				},
			)
		})
	})

	Describe("Webhook validation cases", Ordered, func() {
		var (
			processor         state.ChangeProcessor
			fakeEventRecorder *record.FakeRecorder

			gc *v1beta1.GatewayClass

			gwNsName, hrNsName types.NamespacedName
			gw, gwInvalid      *v1beta1.Gateway
			hr, hrInvalid      *v1beta1.HTTPRoute
		)
		BeforeAll(func() {
			fakeEventRecorder = record.NewFakeRecorder(2 /* number of buffered events */)

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      controllerName,
				GatewayClassName:     gcName,
				RelationshipCapturer: relationship.NewCapturerImpl(gcName),
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
				fakeEventRecorder = record.NewFakeRecorder(1 /* number of buffered events */)

				processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
					GatewayCtlrName:      controllerName,
					GatewayClassName:     gcName,
					RelationshipCapturer: relationship.NewCapturerImpl(gcName),
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
				Entry("listener hostnames conflict",
					createInvalidGateway(func(gw *v1beta1.Gateway) {
						gw.Spec.Listeners = append(gw.Spec.Listeners, v1beta1.Listener{
							Name:     "listener-80-2",
							Hostname: nil,
							Port:     80,
							Protocol: v1beta1.HTTPProtocolType,
						})
					}),
				),
			)
		})
	})

	Describe("Edge cases with panic", func() {
		var (
			processor                state.ChangeProcessor
			fakeRelationshipCapturer *relationshipfakes.FakeCapturer
		)

		BeforeEach(func() {
			fakeRelationshipCapturer = &relationshipfakes.FakeCapturer{}

			processor = state.NewChangeProcessorImpl(state.ChangeProcessorConfig{
				GatewayCtlrName:      "test.controller",
				GatewayClassName:     "my-class",
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
