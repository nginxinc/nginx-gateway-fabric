package telemetry_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events/eventsfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

func createListCallsFunc(nodes []v1.Node) func(
	ctx context.Context,
	list client.ObjectList,
	option ...client.ListOption,
) error {
	return func(_ context.Context, list client.ObjectList, option ...client.ListOption) error {
		Expect(option).To(BeEmpty())

		switch typedList := list.(type) {
		case *v1.NodeList:
			typedList.Items = append(typedList.Items, nodes...)
		default:
			Fail(fmt.Sprintf("unknown type: %T", typedList))
		}
		return nil
	}
}

type getCallsFunc = func(
	context.Context,
	types.NamespacedName,
	client.Object,
	...client.GetOption,
) error

func createGetCallsFunc(objects ...client.Object) getCallsFunc {
	return func(_ context.Context, _ types.NamespacedName, object client.Object, option ...client.GetOption) error {
		Expect(option).To(BeEmpty())

		for _, obj := range objects {
			if reflect.TypeOf(obj) == reflect.TypeOf(object) {
				reflect.ValueOf(object).Elem().Set(reflect.ValueOf(obj).Elem())
				return nil
			}
		}

		return nil
	}
}

var _ = Describe("Collector", Ordered, func() {
	var (
		k8sClientReader         *eventsfakes.FakeReader
		fakeGraphGetter         *telemetryfakes.FakeGraphGetter
		fakeConfigurationGetter *telemetryfakes.FakeConfigurationGetter
		dataCollector           telemetry.DataCollector
		version                 string
		expData                 telemetry.Data
		ctx                     context.Context
		podNSName               types.NamespacedName
		ngfPod                  *v1.Pod
		ngfReplicaSet           *appsv1.ReplicaSet
		kubeNamespace           *v1.Namespace
		baseGetCalls            getCallsFunc
		flags                   config.Flags
	)

	BeforeAll(func() {
		ctx = context.Background()
		version = "1.1"

		ngfPod = &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "ReplicaSet",
						Name: "replicaset1",
					},
				},
			},
		}

		replicas := int32(1)
		ngfReplicaSet = &appsv1.ReplicaSet{
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "replica",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
						Name: "Deployment1",
						UID:  "test-uid-replicaSet",
					},
				},
			},
		}

		podNSName = types.NamespacedName{
			Namespace: "nginx-gateway",
			Name:      "ngf-pod",
		}

		kubeNamespace = &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: metav1.NamespaceSystem,
				UID:  "test-uid",
			},
		}

		flags = config.Flags{
			Names:  []string{"boolFlag", "intFlag", "stringFlag"},
			Values: []string{"false", "default", "user-defined"},
		}
	})

	BeforeEach(func() {
		expData = telemetry.Data{
			ProjectMetadata:   telemetry.ProjectMetadata{Name: "NGF", Version: version},
			NodeCount:         0,
			NGFResourceCounts: telemetry.NGFResourceCounts{},
			NGFReplicaCount:   1,
			ClusterID:         string(kubeNamespace.GetUID()),
<<<<<<< HEAD
			ImageSource:       "local",
			Arch:              runtime.GOARCH,
			DeploymentID:      string(ngfReplicaSet.ObjectMeta.OwnerReferences[0].UID),
			FlagKeyValues:     flagKeyValues,
=======
			Flags:             flags,
>>>>>>> 9e04b8c (Add some review feedback)
		}

		k8sClientReader = &eventsfakes.FakeReader{}
		fakeGraphGetter = &telemetryfakes.FakeGraphGetter{}
		fakeConfigurationGetter = &telemetryfakes.FakeConfigurationGetter{}

		fakeGraphGetter.GetLatestGraphReturns(&graph.Graph{})
		fakeConfigurationGetter.GetLatestConfigurationReturns(&dataplane.Configuration{})

		dataCollector = telemetry.NewDataCollectorImpl(telemetry.DataCollectorConfig{
			K8sClientReader:     k8sClientReader,
			GraphGetter:         fakeGraphGetter,
			ConfigurationGetter: fakeConfigurationGetter,
			Version:             version,
			PodNSName:           podNSName,
<<<<<<< HEAD
			ImageSource:         "local",
			FlagKeyValues:       flagKeyValues,
=======
			Flags:               flags,
>>>>>>> 9e04b8c (Add some review feedback)
		})

		baseGetCalls = createGetCallsFunc(ngfPod, ngfReplicaSet, kubeNamespace)
		k8sClientReader.GetCalls(baseGetCalls)
	})

	mergeGetCallsWithBase := func(f getCallsFunc) getCallsFunc {
		return func(
			ctx context.Context,
			nsName types.NamespacedName,
			object client.Object,
			option ...client.GetOption,
		) error {
			err := baseGetCalls(ctx, nsName, object, option...)
			Expect(err).ToNot(HaveOccurred())

			return f(ctx, nsName, object, option...)
		}
	}

	Describe("Normal case", func() {
		When("collecting telemetry data", func() {
			It("collects all fields", func() {
				nodes := []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node2",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node3",
						},
					},
				}

				k8sClientReader.ListCalls(createListCallsFunc(nodes))

				secret1 := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret1"}}
				secret2 := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret2"}}
				nilsecret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "nilsecret"}}

				svc1 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}}
				svc2 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc2"}}
				nilsvc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "nilsvc"}}

				graph := &graph.Graph{
					GatewayClass: &graph.GatewayClass{},
					Gateway:      &graph.Gateway{},
					IgnoredGatewayClasses: map[types.NamespacedName]*gatewayv1.GatewayClass{
						{Name: "ignoredGC1"}: {},
						{Name: "ignoredGC2"}: {},
					},
					IgnoredGateways: map[types.NamespacedName]*gatewayv1.Gateway{
						{Name: "ignoredGw1"}: {},
						{Name: "ignoredGw2"}: {},
					},
					Routes: map[types.NamespacedName]*graph.Route{
						{Namespace: "test", Name: "hr-1"}: {},
						{Namespace: "test", Name: "hr-2"}: {},
						{Namespace: "test", Name: "hr-3"}: {},
					},
					ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
						client.ObjectKeyFromObject(secret1): {
							Source: secret1,
						},
						client.ObjectKeyFromObject(secret2): {
							Source: secret2,
						},
						client.ObjectKeyFromObject(nilsecret): nil,
					},
					ReferencedServices: map[types.NamespacedName]struct{}{
						client.ObjectKeyFromObject(svc1):   {},
						client.ObjectKeyFromObject(svc2):   {},
						client.ObjectKeyFromObject(nilsvc): {},
					},
				}

				config := &dataplane.Configuration{
					Upstreams: []dataplane.Upstream{
						{
							Name:     "upstream1",
							ErrorMsg: "",
							Endpoints: []resolver.Endpoint{
								{
									Address: "endpoint1",
									Port:    80,
								}, {
									Address: "endpoint2",
									Port:    80,
								}, {
									Address: "endpoint3",
									Port:    80,
								},
							},
						},
						{
							Name:     "upstream2",
							ErrorMsg: "",
							Endpoints: []resolver.Endpoint{
								{
									Address: "endpoint1",
									Port:    80,
								},
							},
						},
					},
				}

				fakeGraphGetter.GetLatestGraphReturns(graph)
				fakeConfigurationGetter.GetLatestConfigurationReturns(config)

				expData.NodeCount = 3
				expData.NGFResourceCounts = telemetry.NGFResourceCounts{
					Gateways:       3,
					GatewayClasses: 3,
					HTTPRoutes:     3,
					Secrets:        3,
					Services:       3,
					Endpoints:      4,
				}

				data, err := dataCollector.Collect(ctx)

				Expect(err).To(BeNil())
				Expect(expData).To(Equal(data))
			})
		})
	})

	Describe("clusterID collector", func() {
		When("collecting clusterID", func() {
			When("it encounters an error while collecting data", func() {
				It("should error if the kubernetes client errored when getting the namespace", func() {
					expectedError := errors.New("there was an error getting clusterID")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(
						func(_ context.Context, _ types.NamespacedName, object client.Object, _ ...client.GetOption) error {
							switch object.(type) {
							case *v1.Namespace:
								return expectedError
							}
							return nil
						}))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedError))
				})
			})
		})
	})

	Describe("node count collector", func() {
		When("collecting node count data", func() {
			It("collects correct data for no nodes", func() {
				k8sClientReader.ListCalls(createListCallsFunc(nil))

				data, err := dataCollector.Collect(ctx)

				Expect(err).To(BeNil())
				Expect(expData).To(Equal(data))
			})

			It("collects correct data for one node", func() {
				nodes := []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "node1"},
					},
				}

				k8sClientReader.ListCalls(createListCallsFunc(nodes))

				expData.NodeCount = 1

				data, err := dataCollector.Collect(ctx)

				Expect(err).To(BeNil())
				Expect(expData).To(Equal(data))
			})
		})
		When("it encounters an error while collecting data", func() {
			It("should error on kubernetes client api errors", func() {
				expectedError := errors.New("there was an error getting NodeList")
				k8sClientReader.ListReturns(expectedError)

				_, err := dataCollector.Collect(ctx)
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("NGF resource count collector", func() {
		var (
			graph1                          *graph.Graph
			config1, invalidUpstreamsConfig *dataplane.Configuration
		)

		BeforeAll(func() {
			secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret1"}}
			svc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}}

			graph1 = &graph.Graph{
				GatewayClass: &graph.GatewayClass{},
				Gateway:      &graph.Gateway{},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-1"}: {},
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					client.ObjectKeyFromObject(secret): {
						Source: secret,
					},
				},
				ReferencedServices: map[types.NamespacedName]struct{}{
					client.ObjectKeyFromObject(svc): {},
				},
			}

			config1 = &dataplane.Configuration{
				Upstreams: []dataplane.Upstream{
					{
						Name:     "upstream1",
						ErrorMsg: "",
						Endpoints: []resolver.Endpoint{
							{
								Address: "endpoint1",
								Port:    80,
							},
						},
					},
				},
			}

			invalidUpstreamsConfig = &dataplane.Configuration{
				Upstreams: []dataplane.Upstream{
					{
						Name:     "invalidUpstream",
						ErrorMsg: "there is an error here",
						Endpoints: []resolver.Endpoint{
							{
								Address: "endpoint1",
								Port:    80,
							}, {
								Address: "endpoint2",
								Port:    80,
							}, {
								Address: "endpoint3",
								Port:    80,
							},
						},
					},
					{
						Name:      "emptyUpstream",
						ErrorMsg:  "",
						Endpoints: []resolver.Endpoint{},
					},
				},
			}
		})

		When("collecting NGF resource counts", func() {
			It("collects correct data for graph with no resources", func() {
				fakeGraphGetter.GetLatestGraphReturns(&graph.Graph{})
				fakeConfigurationGetter.GetLatestConfigurationReturns(&dataplane.Configuration{})

				expData.NGFResourceCounts = telemetry.NGFResourceCounts{}

				data, err := dataCollector.Collect(ctx)

				Expect(err).To(BeNil())
				Expect(expData).To(Equal(data))
			})

			It("collects correct data for graph with one of each resource", func() {
				fakeGraphGetter.GetLatestGraphReturns(graph1)
				fakeConfigurationGetter.GetLatestConfigurationReturns(config1)

				expData.NGFResourceCounts = telemetry.NGFResourceCounts{
					Gateways:       1,
					GatewayClasses: 1,
					HTTPRoutes:     1,
					Secrets:        1,
					Services:       1,
					Endpoints:      1,
				}

				data, err := dataCollector.Collect(ctx)

				Expect(err).To(BeNil())
				Expect(expData).To(Equal(data))
			})

			It("ignores invalid and empty upstreams", func() {
				fakeGraphGetter.GetLatestGraphReturns(&graph.Graph{})
				fakeConfigurationGetter.GetLatestConfigurationReturns(invalidUpstreamsConfig)
				expData.NGFResourceCounts = telemetry.NGFResourceCounts{
					Gateways:       0,
					GatewayClasses: 0,
					HTTPRoutes:     0,
					Secrets:        0,
					Services:       0,
					Endpoints:      0,
				}

				data, err := dataCollector.Collect(ctx)

				Expect(err).To(BeNil())
				Expect(expData).To(Equal(data))
			})

			When("it encounters an error while collecting data", func() {
				BeforeEach(func() {
					fakeGraphGetter.GetLatestGraphReturns(&graph.Graph{})
					fakeConfigurationGetter.GetLatestConfigurationReturns(&dataplane.Configuration{})
				})
				It("should error on nil latest graph", func() {
					expectedError := errors.New("latest graph cannot be nil")
					fakeGraphGetter.GetLatestGraphReturns(nil)

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedError))
				})

				It("should error on nil latest configuration", func() {
					expectedError := errors.New("latest configuration cannot be nil")
					fakeConfigurationGetter.GetLatestConfigurationReturns(nil)

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedError))
				})
			})
		})
	})

	Describe("NGF replica count collector", func() {
		When("collecting NGF replica count", func() {
			When("it encounters an error while collecting data", func() {
				It("should error if the kubernetes client errored when getting the Pod", func() {
					expectedErr := errors.New("there was an error getting the Pod")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(
						func(_ context.Context, _ client.ObjectKey, object client.Object, _ ...client.GetOption) error {
							switch object.(type) {
							case *v1.Pod:
								return expectedErr
							}
							return nil
						},
					))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})

				It("should error if the Pod's owner reference is nil", func() {
					expectedErr := errors.New("expected one owner reference of the NGF Pod, got 0")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:            "pod1",
								OwnerReferences: nil,
							},
						},
					)))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})

				It("should error if the Pod has multiple owner references", func() {
					expectedErr := errors.New("expected one owner reference of the NGF Pod, got 2")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
								OwnerReferences: []metav1.OwnerReference{
									{
										Kind: "ReplicaSet",
										Name: "replicaset1",
									},
									{
										Kind: "ReplicaSet",
										Name: "replicaset2",
									},
								},
							},
						},
					)))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})

				It("should error if the Pod's owner reference is not a ReplicaSet", func() {
					expectedErr := errors.New("expected pod owner reference to be ReplicaSet, got Deployment")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
								OwnerReferences: []metav1.OwnerReference{
									{
										Kind: "Deployment",
										Name: "deployment1",
										UID:  "replica-uid",
									},
								},
							},
						},
					)))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})

				It("should error if the replica set's replicas is nil", func() {
					expectedErr := errors.New("replica set replicas was nil")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&appsv1.ReplicaSet{
							Spec: appsv1.ReplicaSetSpec{
								Replicas: nil,
							},
						},
					)))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})

				It("should error if the kubernetes client errored when getting the ReplicaSet", func() {
					expectedErr := errors.New("there was an error getting the ReplicaSet")
					k8sClientReader.GetCalls(mergeGetCallsWithBase(
						func(_ context.Context, _ client.ObjectKey, object client.Object, _ ...client.GetOption) error {
							switch object.(type) {
							case *appsv1.ReplicaSet:
								return expectedErr
							}
							return nil
						}))

					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})
	})

	Describe("DeploymentID collector", func() {
		When("collecting deploymentID", func() {
			When("it encounters an error while collecting data", func() {
				It("should error if the replicaSet's owner reference is nil", func() {
					replicas := int32(1)
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&appsv1.ReplicaSet{
							Spec: appsv1.ReplicaSetSpec{
								Replicas: &replicas,
							},
						},
					)))

					expectedErr := errors.New("expected one owner reference of the NGF ReplicaSet, got 0")
					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})

				It("should error if the replicaSet's owner reference kind is not deployment", func() {
					replicas := int32(1)
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&appsv1.ReplicaSet{
							Spec: appsv1.ReplicaSetSpec{
								Replicas: &replicas,
							},
							ObjectMeta: metav1.ObjectMeta{
								OwnerReferences: []metav1.OwnerReference{
									{
										Name: "replica",
										Kind: "ReplicaSet",
									},
								},
							},
						},
					)))

					expectedErr := errors.New("expected replicaSet owner reference to be Deployment, got ReplicaSet")
					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})
				It("should error if the replicaSet's owner reference has empty UID", func() {
					replicas := int32(1)
					k8sClientReader.GetCalls(mergeGetCallsWithBase(createGetCallsFunc(
						&appsv1.ReplicaSet{
							Spec: appsv1.ReplicaSetSpec{
								Replicas: &replicas,
							},
							ObjectMeta: metav1.ObjectMeta{
								OwnerReferences: []metav1.OwnerReference{
									{
										Name: "replica",
										Kind: "Deployment",
									},
								},
							},
						},
					)))

					expectedErr := errors.New("expected replicaSet owner reference to have a UID")
					_, err := dataCollector.Collect(ctx)
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})
	})
})
