package state_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

const gatewayCtlrName = v1alpha2.GatewayController("test-name")

var _ = Describe("Configuration", func() {
	Describe("Processing HTTPRoutes", func() {
		var conf state.Configuration

		constTime := time.Now()

		BeforeEach(OncePerOrdered, func() {
			conf = state.NewConfiguration(string(gatewayCtlrName), state.NewFakeClock(constTime))
		})

		Describe("Process one HTTPRoute with one host", Ordered, func() {
			var hr, updatedHRWithSameGen, updatedHRWithIncrementedGen *v1alpha2.HTTPRoute

			BeforeAll(func() {
				hr = &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "route1",
					},
					Spec: v1alpha2.HTTPRouteSpec{
						Hostnames: []v1alpha2.Hostname{
							"cafe.example.com",
						},
						Rules: []v1alpha2.HTTPRouteRule{
							{
								// mo matches -> "/"
							},
							{
								Matches: []v1alpha2.HTTPRouteMatch{
									{
										Path: &v1alpha2.HTTPPathMatch{
											Value: helpers.GetStringPointer("/coffee"),
										},
									},
								},
							},
						},
					},
				}

				updatedHRWithSameGen = hr.DeepCopy()
				updatedHRWithSameGen.Spec.Rules[1].Matches[0].Path.Value = helpers.GetStringPointer("/tea")

				updatedHRWithIncrementedGen = updatedHRWithSameGen.DeepCopy()
				updatedHRWithIncrementedGen.Generation++
			})

			It("should upsert a host and generate a status update for the new HTTPRoute", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr,
										},
									},
								},
								{
									Path: "/coffee",
									Routes: []state.Route{
										{
											MatchIdx: 0,
											RuleIdx:  1,
											Source:   hr,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should not generate changes and status updates for the updated HTTPRoute because it has the same generation as the old one", func() {
				changes, statusUpdates := conf.UpsertHTTPRoute(updatedHRWithSameGen)
				Expect(changes).To(BeEmpty())
				Expect(statusUpdates).To(BeEmpty())
			})

			It("should upsert the host changes and generate a status update for the updated HTTPRoute", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   updatedHRWithIncrementedGen,
										},
									},
								},
								{
									Path: "/tea",
									Routes: []state.Route{
										{
											MatchIdx: 0,
											RuleIdx:  1,
											Source:   updatedHRWithIncrementedGen,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: updatedHRWithIncrementedGen.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(updatedHRWithIncrementedGen)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should delete the host for the deleted HTTPRoute", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Delete,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   updatedHRWithIncrementedGen,
										},
									},
								},
								{
									Path: "/tea",
									Routes: []state.Route{
										{
											MatchIdx: 0,
											RuleIdx:  1,
											Source:   updatedHRWithIncrementedGen,
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.DeleteHTTPRoute(types.NamespacedName{Namespace: "test", Name: "route1"})
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(statusUpdates).To(BeEmpty())
			})
		})

		It("should allow removing an non-exiting HTTPRoute", func() {
			changes, statusUpdates := conf.DeleteHTTPRoute(types.NamespacedName{Namespace: "test", Name: "some-route"})
			Expect(changes).To(BeEmpty())
			Expect(statusUpdates).To(BeEmpty())
		})

		Describe("Processing multiple HTTPRoutes for one host", Ordered, func() {
			var hr1, hr2, hr2Updated *v1alpha2.HTTPRoute

			BeforeAll(func() {
				hr1 = &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "route1",
					},
					Spec: v1alpha2.HTTPRouteSpec{
						Hostnames: []v1alpha2.Hostname{
							"cafe.example.com",
						},
						Rules: []v1alpha2.HTTPRouteRule{
							{
								// mo matches -> "/"
							},
						},
					},
				}

				hr2 = &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "route2",
					},
					Spec: v1alpha2.HTTPRouteSpec{
						Hostnames: []v1alpha2.Hostname{
							"cafe.example.com",
						},
						Rules: []v1alpha2.HTTPRouteRule{
							{
								Matches: []v1alpha2.HTTPRouteMatch{
									{
										Path: &v1alpha2.HTTPPathMatch{
											Value: helpers.GetStringPointer("/coffee"),
										},
									},
								},
							},
						},
					},
				}

				hr2Updated = hr2.DeepCopy()
				hr2Updated.Spec.Rules[0].Matches[0].Path.Value = helpers.GetStringPointer("/tea")
				hr2Updated.Generation++
			})

			It("should upsert a host and generate a status update for the first HTTPRoute", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr1)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should upsert the same host and generate status updates for both HTTPRoutes after adding the second", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
								{
									Path: "/coffee",
									Routes: []state.Route{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr2,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route2"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr2.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr2)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should upsert the host and generate status updates for both HTTPRoutes after updating the second", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
								{
									Path: "/tea",
									Routes: []state.Route{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr2Updated,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route2"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr2Updated.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr2Updated)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should upsert the host and generate a status updates for the first HTTPRoute after deleting the second", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.DeleteHTTPRoute(types.NamespacedName{Namespace: "test", Name: "route2"})
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should delete the host after deleting the first HTTPRoute", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Delete,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
							},
						},
					},
				}
				changes, statusUpdates := conf.DeleteHTTPRoute(types.NamespacedName{Namespace: "test", Name: "route1"})
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(statusUpdates).To(BeEmpty())
			})
		})

		Describe("Processing conflicting HTTPRoutes", Ordered, func() {
			var earlier, later metav1.Time
			var hr1, hr2, hr1Updated *v1alpha2.HTTPRoute

			BeforeAll(func() {
				earlier = metav1.Now()
				later = metav1.NewTime(earlier.Add(1 * time.Second))

				hr1 = &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:         "test",
						Name:              "route1",
						CreationTimestamp: earlier,
					},
					Spec: v1alpha2.HTTPRouteSpec{
						Hostnames: []v1alpha2.Hostname{
							"cafe.example.com",
						},
						Rules: []v1alpha2.HTTPRouteRule{
							{
								// mo matches -> "/"
							},
						},
					},
				}

				hr2 = &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:         "test",
						Name:              "route2",
						CreationTimestamp: earlier,
					},
					Spec: v1alpha2.HTTPRouteSpec{
						Hostnames: []v1alpha2.Hostname{
							"cafe.example.com",
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

				hr1Updated = hr1.DeepCopy()
				hr1Updated.Generation++
				hr1Updated.CreationTimestamp = later
			})

			It("should upsert a host and generate a status update for the first HTTPRoute", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr1)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should upsert the host (make the first HTTPRoute the winner for '/' rule) and generate status updates for both HTTPRoutes after adding the second", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1,
										},
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr2,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route2"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr2.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr2)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})

			It("should upsert the host (make the second HTTPRoute the winner for '/' rule) and generate status updates for both HTTPRoutes after updating the first", func() {
				expectedChanges := []state.Change{
					{
						Op: state.Upsert,
						Host: state.Host{
							Value: "cafe.example.com",
							PathRouteGroups: []state.PathRouteGroup{
								{
									Path: "/",
									Routes: []state.Route{
										{
											MatchIdx: 0,
											RuleIdx:  0,
											Source:   hr2,
										},
										{
											MatchIdx: -1,
											RuleIdx:  0,
											Source:   hr1Updated,
										},
									},
								},
							},
						},
					},
				}
				expectedStatusUpdates := []state.StatusUpdate{
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr1Updated.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
					{
						NamespacedName: types.NamespacedName{Namespace: "test", Name: "route2"},
						Status: &v1alpha2.HTTPRouteStatus{
							RouteStatus: v1alpha2.RouteStatus{
								Parents: []v1alpha2.RouteParentStatus{
									{
										ControllerName: gatewayCtlrName,
										ParentRef: v1alpha2.ParentRef{
											Name: "fake",
										},
										Conditions: []metav1.Condition{
											{
												Type:               string(v1alpha2.ConditionRouteAccepted),
												Status:             "True",
												ObservedGeneration: hr2.Generation,
												LastTransitionTime: metav1.NewTime(constTime),
												Reason:             string(v1alpha2.ConditionRouteAccepted),
												Message:            "",
											},
										},
									},
								},
							},
						},
					},
				}

				changes, statusUpdates := conf.UpsertHTTPRoute(hr1Updated)
				Expect(helpers.Diff(expectedChanges, changes)).To(BeEmpty())
				Expect(helpers.Diff(expectedStatusUpdates, statusUpdates)).To(BeEmpty())
			})
		})
	})
})

func TestRouteGetMatch(t *testing.T) {
	var hr = &v1alpha2.HTTPRoute{
		Spec: v1alpha2.HTTPRouteSpec{
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-1"),
							},
						},
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-2"),
							},
						},
					},
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-3"),
							},
						},
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-4"),
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name,
		expPath string
		route       state.Route
		matchExists bool
	}{
		{
			name:        "match does not exist",
			expPath:     "",
			route:       state.Route{MatchIdx: -1},
			matchExists: false,
		},
		{
			name:        "first match in first rule",
			expPath:     "/path-1",
			route:       state.Route{MatchIdx: 0, RuleIdx: 0, Source: hr},
			matchExists: true,
		},
		{
			name:        "second match in first rule",
			expPath:     "/path-2",
			route:       state.Route{MatchIdx: 1, RuleIdx: 0, Source: hr},
			matchExists: true,
		},
		{
			name:        "second match in second rule",
			expPath:     "/path-4",
			route:       state.Route{MatchIdx: 1, RuleIdx: 1, Source: hr},
			matchExists: true,
		},
	}

	for _, tc := range tests {
		actual, exists := tc.route.GetMatch()
		if !tc.matchExists {
			if exists {
				t.Errorf("route.GetMatch() incorrectly returned true (match exists) for test case: %q", tc.name)
			}
		} else {
			if !exists {
				t.Errorf("route.GetMatch() incorrectly returned false (match does not exist) for test case: %q", tc.name)
			}
			if *actual.Path.Value != tc.expPath {
				t.Errorf("route.GetMatch() returned incorrect match with path: %s, expected path: %s for test case: %q", *actual.Path.Value, tc.expPath, tc.name)
			}
		}
	}
}
