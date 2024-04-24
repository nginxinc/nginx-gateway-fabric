package clientsettings_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/clientsettings"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/policiesfakes"
)

type policyModFunc func(policy *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy

func createValidPolicy() *ngfAPI.ClientSettingsPolicy {
	return &ngfAPI.ClientSettingsPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: ngfAPI.ClientSettingsPolicySpec{
			TargetRef: v1alpha2.LocalPolicyTargetReference{
				Group: v1.GroupName,
				Kind:  kinds.Gateway,
				Name:  "gateway",
			},
			Body: &ngfAPI.ClientBody{
				MaxSize: helpers.GetPointer[ngfAPI.Size]("10m"),
				Timeout: helpers.GetPointer[ngfAPI.Duration]("600ms"),
			},
			KeepAlive: &ngfAPI.ClientKeepAlive{
				Requests: helpers.GetPointer[int32](900),
				Time:     helpers.GetPointer[ngfAPI.Duration]("50s"),
				Timeout: &ngfAPI.ClientKeepAliveTimeout{
					Server: helpers.GetPointer[ngfAPI.Duration]("30s"),
					Header: helpers.GetPointer[ngfAPI.Duration]("60s"),
				},
			},
		},
		Status: v1alpha2.PolicyStatus{},
	}
}

func createModifiedPolicy(mod policyModFunc) *ngfAPI.ClientSettingsPolicy {
	return mod(createValidPolicy())
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name             string
		policy           *ngfAPI.ClientSettingsPolicy
		expErrSubstrings []string
	}{
		{
			name: "invalid target ref; unsupported group",
			policy: createModifiedPolicy(func(p *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy {
				p.Spec.TargetRef.Group = "Unsupported"
				return p
			}),
			expErrSubstrings: []string{"spec.targetRef.group"},
		},
		{
			name: "invalid target ref; unsupported kind",
			policy: createModifiedPolicy(func(p *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy {
				p.Spec.TargetRef.Kind = "Unsupported"
				return p
			}),
			expErrSubstrings: []string{"spec.targetRef.kind"},
		},
		{
			name: "invalid client max body size",
			policy: createModifiedPolicy(func(p *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy {
				p.Spec.Body.MaxSize = helpers.GetPointer[ngfAPI.Size]("invalid")
				return p
			}),
			expErrSubstrings: []string{"spec.body.maxSize"},
		},
		{
			name: "invalid durations",
			policy: createModifiedPolicy(func(p *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy {
				p.Spec.Body.Timeout = helpers.GetPointer[ngfAPI.Duration]("invalid")
				p.Spec.KeepAlive.Time = helpers.GetPointer[ngfAPI.Duration]("invalid")
				p.Spec.KeepAlive.Timeout.Server = helpers.GetPointer[ngfAPI.Duration]("invalid")
				p.Spec.KeepAlive.Timeout.Header = helpers.GetPointer[ngfAPI.Duration]("invalid")
				return p
			}),
			expErrSubstrings: []string{
				"spec.body.timeout",
				"spec.keepAlive.time",
				"spec.keepAlive.timeout.server",
				"spec.keepAlive.timeout.header",
			},
		},
		{
			name: "invalid keepalive requests",
			policy: createModifiedPolicy(func(p *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy {
				p.Spec.KeepAlive.Requests = helpers.GetPointer[int32](-1)
				return p
			}),
			expErrSubstrings: []string{"spec.keepAlive.requests"},
		},
		{
			name: "invalid keepalive timeout; header provided without server",
			policy: createModifiedPolicy(func(p *ngfAPI.ClientSettingsPolicy) *ngfAPI.ClientSettingsPolicy {
				p.Spec.KeepAlive.Timeout.Server = nil
				return p
			}),
			expErrSubstrings: []string{"spec.keepAlive.timeout"},
		},
		{
			name:             "valid",
			policy:           createValidPolicy(),
			expErrSubstrings: nil,
		},
	}

	v := clientsettings.NewValidator(validation.GenericValidator{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := v.Validate(test.policy)

			if len(test.expErrSubstrings) == 0 {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(HaveOccurred())
			}

			for _, str := range test.expErrSubstrings {
				g.Expect(err.Error()).To(ContainSubstring(str))
			}
		})
	}
}

func TestValidator_ValidatePanics(t *testing.T) {
	v := clientsettings.NewValidator(nil)

	validate := func() {
		_ = v.Validate(&policiesfakes.FakePolicy{})
	}

	g := NewWithT(t)

	g.Expect(validate).To(Panic())
}

func TestValidator_Conflicts(t *testing.T) {
	tests := []struct {
		polA      *ngfAPI.ClientSettingsPolicy
		polB      *ngfAPI.ClientSettingsPolicy
		name      string
		conflicts bool
	}{
		{
			name: "no conflicts",
			polA: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						MaxSize: helpers.GetPointer[ngfAPI.Size]("10m"),
					},
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Requests: helpers.GetPointer[int32](900),
						Time:     helpers.GetPointer[ngfAPI.Duration]("50s"),
					},
				},
			},
			polB: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						Timeout: helpers.GetPointer[ngfAPI.Duration]("600ms"),
					},
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Timeout: &ngfAPI.ClientKeepAliveTimeout{
							Server: helpers.GetPointer[ngfAPI.Duration]("30s"),
							Header: helpers.GetPointer[ngfAPI.Duration]("60s"),
						},
					},
				},
			},
			conflicts: false,
		},
		{
			name: "body max size conflicts",
			polA: createValidPolicy(),
			polB: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						MaxSize: helpers.GetPointer[ngfAPI.Size]("10m"),
					},
				},
			},
			conflicts: true,
		},
		{
			name: "body timeout conflicts",
			polA: createValidPolicy(),
			polB: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						Timeout: helpers.GetPointer[ngfAPI.Duration]("600ms"),
					},
				},
			},
			conflicts: true,
		},
		{
			name: "keepalive requests conflicts",
			polA: createValidPolicy(),
			polB: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Requests: helpers.GetPointer[int32](900),
					},
				},
			},
			conflicts: true,
		},
		{
			name: "keepalive time conflicts",
			polA: createValidPolicy(),
			polB: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Time: helpers.GetPointer[ngfAPI.Duration]("50s"),
					},
				},
			},
			conflicts: true,
		},
		{
			name: "keepalive timeout conflicts",
			polA: createValidPolicy(),
			polB: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Timeout: &ngfAPI.ClientKeepAliveTimeout{
							Server: helpers.GetPointer[ngfAPI.Duration]("30s"),
						},
					},
				},
			},
			conflicts: true,
		},
	}

	v := clientsettings.NewValidator(nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(v.Conflicts(test.polA, test.polB)).To(Equal(test.conflicts))
		})
	}
}

func TestValidator_ConflictsPanics(t *testing.T) {
	v := clientsettings.NewValidator(nil)

	conflicts := func() {
		_ = v.Conflicts(&policiesfakes.FakePolicy{}, &policiesfakes.FakePolicy{})
	}

	g := NewWithT(t)

	g.Expect(conflicts).To(Panic())
}
