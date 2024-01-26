package telemetry_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events/eventsfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

var _ = Describe("Collector", Ordered, func() {
	var (
		k8sClientReader *eventsfakes.FakeReader
		fakeGraphGetter *telemetryfakes.FakeGraphGetter
		dataCollector   telemetry.DataCollector
		version         string
		graph1, graph2  *graph.Graph
		ctx             context.Context
	)

	BeforeAll(func() {
		ctx = context.Background()
		version = "1.1"

		k8sClientReader = &eventsfakes.FakeReader{}
		fakeGraphGetter = &telemetryfakes.FakeGraphGetter{}
		fakeGraphGetter.GetLatestGraphReturns(&graph.Graph{})

		dataCollector = telemetry.NewDataCollector(telemetry.DataCollectorConfig{
			K8sClientReader: k8sClientReader,
			GraphGetter:     fakeGraphGetter,
			Version:         version,
		})

		secret1 := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret1"}}
		secret2 := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "secret2"}}
		nilsecret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "nilsecret"}}

		svc1 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}}
		svc2 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc2"}}
		nilsvc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "nilsvc"}}

		graph1 = &graph.Graph{
			GatewayClass: &graph.GatewayClass{},
			Gateway:      &graph.Gateway{},
			Routes: map[types.NamespacedName]*graph.Route{
				{Namespace: "test", Name: "hr-1"}: {},
			},
			ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
				client.ObjectKeyFromObject(secret1): {
					Source: secret1,
				},
			},
			ReferencedServices: map[types.NamespacedName]*v1.Service{
				client.ObjectKeyFromObject(svc1): svc1,
			},
		}
		graph2 = &graph.Graph{
			GatewayClass: &graph.GatewayClass{},
			Gateway:      &graph.Gateway{},
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
			ReferencedServices: map[types.NamespacedName]*v1.Service{
				client.ObjectKeyFromObject(svc1):   svc1,
				client.ObjectKeyFromObject(svc2):   svc2,
				client.ObjectKeyFromObject(nilsvc): nil,
			},
		}
	})

	When("retrieving node count data", func() {
		It("generates correct data for no nodes", func() {
			k8sClientReader.ListCalls(func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
				Expect(option).To(BeEmpty())

				switch typedList := list.(type) {
				case *v1.NodeList:
					typedList.Items = []v1.Node{}
				default:
					Fail(fmt.Sprintf("unknown type: %T", typedList))
				}
				return nil
			})

			expData := telemetry.Data{
				ProjectMetadata:   telemetry.ProjectMetadata{Name: "NGF", Version: version},
				NodeCount:         0,
				NGFResourceCounts: telemetry.NGFResourceCounts{},
			}

			data, err := dataCollector.Collect(ctx)

			Expect(err).To(BeNil())
			Expect(expData).To(Equal(data))
		})

		It("generates correct data for one node", func() {
			node := v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			}

			k8sClientReader.ListCalls(func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
				Expect(option).To(BeEmpty())

				switch typedList := list.(type) {
				case *v1.NodeList:
					typedList.Items = append(typedList.Items, node)
				default:
					Fail(fmt.Sprintf("unknown type: %T", typedList))
				}
				return nil
			})

			expData := telemetry.Data{
				ProjectMetadata:   telemetry.ProjectMetadata{Name: "NGF", Version: version},
				NodeCount:         1,
				NGFResourceCounts: telemetry.NGFResourceCounts{},
			}

			data, err := dataCollector.Collect(ctx)

			Expect(err).To(BeNil())
			Expect(expData).To(Equal(data))
		})

		It("generates correct data for multiple nodes", func() {
			node := v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			}
			node2 := v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node2"},
			}
			node3 := v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node3"},
			}

			k8sClientReader.ListCalls(func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
				Expect(option).To(BeEmpty())

				switch typedList := list.(type) {
				case *v1.NodeList:
					typedList.Items = append(typedList.Items, node, node2, node3)
				default:
					Fail(fmt.Sprintf("unknown type: %T", typedList))
				}
				return nil
			})

			expData := telemetry.Data{
				ProjectMetadata:   telemetry.ProjectMetadata{Name: "NGF", Version: version},
				NodeCount:         3,
				NGFResourceCounts: telemetry.NGFResourceCounts{},
			}

			data, err := dataCollector.Collect(ctx)

			Expect(err).To(BeNil())
			Expect(expData).To(Equal(data))
		})
	})

	When("retrieving NGF resource counts", func() {
		It("generates correct data for graph with one of each resource", func() {
			fakeGraphGetter.GetLatestGraphReturns(graph1)

			expData := telemetry.Data{
				ProjectMetadata: telemetry.ProjectMetadata{Name: "NGF", Version: version},
				NodeCount:       3,
				NGFResourceCounts: telemetry.NGFResourceCounts{
					Gateways:       1,
					GatewayClasses: 1,
					HTTPRoutes:     1,
					Secrets:        1,
					Services:       1,
				},
			}

			data, err := dataCollector.Collect(ctx)

			Expect(err).To(BeNil())
			Expect(expData).To(Equal(data))
		})

		It("generates correct data for graph with multiple of each resource", func() {
			fakeGraphGetter.GetLatestGraphReturns(graph2)

			expData := telemetry.Data{
				ProjectMetadata: telemetry.ProjectMetadata{Name: "NGF", Version: version},
				NodeCount:       3,
				NGFResourceCounts: telemetry.NGFResourceCounts{
					Gateways:       1,
					GatewayClasses: 1,
					HTTPRoutes:     3,
					Secrets:        2,
					Services:       2,
				},
			}

			data, err := dataCollector.Collect(ctx)

			Expect(err).To(BeNil())
			Expect(expData).To(Equal(data))
		})
	})

	When("it encounters an error while collecting data", func() {
		It("should error on client errors", func() {
			k8sClientReader.ListReturns(errors.New("there was an error"))

			_, err := dataCollector.Collect(ctx)
			Expect(err).To(HaveOccurred())
		})
	})
})
