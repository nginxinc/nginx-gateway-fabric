package newstate_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
)

var _ = Describe("ChangeProcessor", func() {
	Describe("Normal cases of processing changes", func() {
		var (
			hr1, hr1Updated, hr2 *v1alpha2.HTTPRoute
			gw, gwUpdated        *v1alpha2.Gateway
			processor            newstate.ChangeProcessor
		)

		BeforeEach(OncePerOrdered, func() {
			createRoute := func(name string, hostname string) *v1alpha2.HTTPRoute {
				return &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      name,
					},
					Spec: v1alpha2.HTTPRouteSpec{
						CommonRouteSpec: v1alpha2.CommonRouteSpec{
							ParentRefs: []v1alpha2.ParentRef{
								{
									Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
									Name:        "gateway",
									SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("listener-80-1")),
								},
							},
						},
						Hostnames: []v1alpha2.Hostname{
							v1alpha2.Hostname(hostname),
						},
						Rules: []v1alpha2.HTTPRouteRule{
							{
								Matches: []v1alpha2.HTTPRouteMatch{
									{
										Path: &v1alpha2.HTTPPathMatch{
											Value: helpers.GetStringPointer("/"),
										},
									},
								},
							},
						},
					},
				}
			}

			hr1 = createRoute("hr-1", "foo.example.com")

			hr1Updated = hr1.DeepCopy()
			hr1Updated.Generation++

			hr2 = createRoute("hr-2", "bar.example.com")

			gw = &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "gateway",
				},
				Spec: v1alpha2.GatewaySpec{
					Listeners: []v1alpha2.Listener{
						{
							Name:     "listener-80-1",
							Hostname: nil,
							Port:     80,
							Protocol: v1alpha2.HTTPProtocolType,
						},
					},
				},
			}

			gwUpdated = gw.DeepCopy()
			gwUpdated.Generation++

			processor = newstate.NewChangeProcessorImpl(types.NamespacedName{Namespace: "test", Name: "gateway"})
		})

		Describe("Process resources", Ordered, func() {
			It("should return empty configuration and statuses when no upsert has occurred", func() {
				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeFalse())
				Expect(conf).To(BeZero())
				Expect(statuses).To(BeZero())
			})

			It("should return empty configuration and updated statuses after upserting an HTTPRoute when the Gateway doesn't exist", func() {
				processor.CaptureUpsertChange(hr1)

				expectedConf := newstate.Configuration{HTTPServers: []newstate.HTTPServer{}}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: false},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return updated configuration and statuses after the Gateway is upserted", func() {
				processor.CaptureUpsertChange(gw)

				expectedConf := newstate.Configuration{
					HTTPServers: []newstate.HTTPServer{
						{
							Hostname: "foo.example.com",
							PathRules: []newstate.PathRule{
								{
									Path: "/",
									MatchRules: []newstate.MatchRule{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
							},
						},
					},
				}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{
						"listener-80-1": {
							Valid:          true,
							AttachedRoutes: 1,
						},
					},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: true},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return empty configuration and statuses after processing upserting the HTTPRoute without generation change", func() {
				hr1UpdatedSameGen := hr1.DeepCopy()
				// hr1UpdatedSameGen.Generation has not been changed
				processor.CaptureUpsertChange(hr1UpdatedSameGen)

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeFalse())
				Expect(conf).To(BeZero())
				Expect(statuses).To(BeZero())
			})

			It("should return updated configuration and statuses after upserting the HTTPRoute with generation change", func() {
				processor.CaptureUpsertChange(hr1Updated)

				expectedConf := newstate.Configuration{
					HTTPServers: []newstate.HTTPServer{
						{
							Hostname: "foo.example.com",
							PathRules: []newstate.PathRule{
								{
									Path: "/",
									MatchRules: []newstate.MatchRule{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr1Updated,
										},
									},
								},
							},
						},
					},
				}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{
						"listener-80-1": {
							Valid:          true,
							AttachedRoutes: 1,
						},
					},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: true},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return empty configuration and statuses after processing upserting the Gateway without generation change", func() {
				gwUpdatedSameGen := gw.DeepCopy()
				// gwUpdatedSameGen.Generation has not been changed
				processor.CaptureUpsertChange(gwUpdatedSameGen)

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeFalse())
				Expect(conf).To(BeZero())
				Expect(statuses).To(BeZero())
			})

			It("should return updated configuration and statuses after upserting the Gateway with generation change", func() {
				processor.CaptureUpsertChange(gwUpdated)

				expectedConf := newstate.Configuration{
					HTTPServers: []newstate.HTTPServer{
						{
							Hostname: "foo.example.com",
							PathRules: []newstate.PathRule{
								{
									Path: "/",
									MatchRules: []newstate.MatchRule{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr1Updated,
										},
									},
								},
							},
						},
					},
				}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{
						"listener-80-1": {
							Valid:          true,
							AttachedRoutes: 1,
						},
					},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: true},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return empty configuration and statuses after processing without capturing any changes", func() {
				changed, conf, statuses := processor.Process()

				Expect(changed).To(BeFalse())
				Expect(conf).To(BeZero())
				Expect(statuses).To(BeZero())
			})

			It("should return updated configuration and statuses after a second HTTPRoute is upserted", func() {
				processor.CaptureUpsertChange(hr2)

				expectedConf := newstate.Configuration{
					HTTPServers: []newstate.HTTPServer{
						{
							Hostname: "bar.example.com",
							PathRules: []newstate.PathRule{
								{
									Path: "/",
									MatchRules: []newstate.MatchRule{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr2,
										},
									},
								},
							},
						},
						{
							Hostname: "foo.example.com",
							PathRules: []newstate.PathRule{
								{
									Path: "/",
									MatchRules: []newstate.MatchRule{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr1Updated,
										},
									},
								},
							},
						},
					},
				}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{
						"listener-80-1": {
							Valid:          true,
							AttachedRoutes: 2,
						},
					},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: true},
							},
						},
						{Namespace: "test", Name: "hr-2"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: true},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return updated configuration and statuses after deleting the second HTTPRoute", func() {
				processor.CaptureDeleteChange(&v1alpha2.HTTPRoute{}, types.NamespacedName{Namespace: "test", Name: "hr-2"})

				expectedConf := newstate.Configuration{
					HTTPServers: []newstate.HTTPServer{
						{
							Hostname: "foo.example.com",
							PathRules: []newstate.PathRule{
								{
									Path: "/",
									MatchRules: []newstate.MatchRule{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr1Updated,
										},
									},
								},
							},
						},
					},
				}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{
						"listener-80-1": {
							Valid:          true,
							AttachedRoutes: 1,
						},
					},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: true},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return empty configuration and updated statuses after deleting the Gateway", func() {
				processor.CaptureDeleteChange(&v1alpha2.Gateway{}, types.NamespacedName{Namespace: "test", Name: "gateway"})

				expectedConf := newstate.Configuration{HTTPServers: []newstate.HTTPServer{}}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "hr-1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"listener-80-1": {Attached: false},
							},
						},
					},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})

			It("should return empty configuration and statuses after deleting the first HTTPRoute", func() {
				processor.CaptureDeleteChange(&v1alpha2.HTTPRoute{}, types.NamespacedName{Namespace: "test", Name: "hr-1"})

				expectedConf := newstate.Configuration{HTTPServers: []newstate.HTTPServer{}}
				expectedStatuses := newstate.Statuses{
					ListenerStatuses:  map[string]newstate.ListenerStatus{},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{},
				}

				changed, conf, statuses := processor.Process()
				Expect(changed).To(BeTrue())
				Expect(helpers.Diff(expectedConf, conf)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatuses, statuses)).To(BeEmpty())
			})
		})
	})

	Describe("Edge cases with panic", func() {
		var processor newstate.ChangeProcessor

		BeforeEach(func() {
			processor = newstate.NewChangeProcessorImpl(types.NamespacedName{Namespace: "test", Name: "gateway"})
		})

		DescribeTable("CaptureUpsertChange must panic",
			func(obj client.Object) {
				process := func() {
					processor.CaptureUpsertChange(obj)
				}
				Expect(process).Should(Panic())
			},
			Entry("an unsupported resource", &v1alpha2.TCPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "tcp"}}),
			Entry("a wrong gateway", &v1alpha2.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "other-gateway"}}))

		DescribeTable("CaptureDeleteChange must panic",
			func(resourceType client.Object, nsname types.NamespacedName) {
				process := func() {
					processor.CaptureDeleteChange(resourceType, nsname)
				}
				Expect(process).Should(Panic())
			},
			Entry("an unsupported resource", &v1alpha2.TCPRoute{}, types.NamespacedName{Namespace: "test", Name: "tcp"}),
			Entry("a wrong gateway", &v1alpha2.Gateway{}, types.NamespacedName{Namespace: "test", Name: "other-gateway"}))
	})
})
