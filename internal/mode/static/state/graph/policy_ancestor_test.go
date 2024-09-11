package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

func TestBackendTLSPolicyAncestorsFull(t *testing.T) {
	t.Parallel()
	createCurStatus := func(numAncestors int, ctlrName string) []v1alpha2.PolicyAncestorStatus {
		statuses := make([]v1alpha2.PolicyAncestorStatus, 0, numAncestors)

		for range numAncestors {
			statuses = append(statuses, v1alpha2.PolicyAncestorStatus{
				ControllerName: v1.GatewayController(ctlrName),
			})
		}

		return statuses
	}

	tests := []struct {
		name      string
		curStatus []v1alpha2.PolicyAncestorStatus
		expFull   bool
	}{
		{
			name:      "not full",
			curStatus: createCurStatus(15, "controller"),
			expFull:   false,
		},
		{
			name:      "full; ancestor does not exist in current status",
			curStatus: createCurStatus(16, "controller"),
			expFull:   true,
		},
		{
			name:      "full, but ancestor does exist in current status",
			curStatus: createCurStatus(16, "nginx-gateway"),
			expFull:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			full := backendTLSPolicyAncestorsFull(test.curStatus, "nginx-gateway")
			g.Expect(full).To(Equal(test.expFull))
		})
	}
}

func TestNGFPolicyAncestorsFull(t *testing.T) {
	t.Parallel()
	type ancestorConfig struct {
		numCurrNGFAncestors    int
		numCurrNonNGFAncestors int
		numNewNGFAncestors     int
	}

	createPolicy := func(cfg ancestorConfig) *Policy {
		currAncestors := make([]v1alpha2.PolicyAncestorStatus, 0, cfg.numCurrNGFAncestors+cfg.numCurrNonNGFAncestors)
		ngfAncestors := make([]PolicyAncestor, 0, cfg.numNewNGFAncestors)

		for range cfg.numCurrNonNGFAncestors {
			currAncestors = append(currAncestors, v1alpha2.PolicyAncestorStatus{
				ControllerName: "non-ngf",
			})
		}

		for range cfg.numCurrNGFAncestors {
			currAncestors = append(currAncestors, v1alpha2.PolicyAncestorStatus{
				ControllerName: "nginx-gateway",
			})
		}

		for range cfg.numNewNGFAncestors {
			ngfAncestors = append(ngfAncestors, PolicyAncestor{
				Ancestor: v1.ParentReference{},
			})
		}

		return &Policy{
			Source: &ngfAPI.ObservabilityPolicy{
				Status: v1alpha2.PolicyStatus{
					Ancestors: currAncestors,
				},
			},
			Ancestors: ngfAncestors,
		}
	}

	tests := []struct {
		name    string
		expFull bool
		cfg     ancestorConfig
	}{
		{
			name: "current policy not full, no new NGF ancestors have been built yet",
			cfg: ancestorConfig{
				numCurrNGFAncestors:    3,
				numCurrNonNGFAncestors: 12,
				numNewNGFAncestors:     0,
			},
			expFull: false,
		},
		{
			name: "current policy not full, and some new NGF ancestors have been built (not at max)",
			cfg: ancestorConfig{
				numCurrNGFAncestors:    3,
				numCurrNonNGFAncestors: 11,
				numNewNGFAncestors:     2,
			},
			expFull: false,
		},
		{
			name: "current policy not full, and some new NGF ancestors have been built (at max)",
			cfg: ancestorConfig{
				numCurrNGFAncestors:    3,
				numCurrNonNGFAncestors: 11,
				numNewNGFAncestors:     5,
			},
			expFull: true,
		},
		{
			name: "current policy is full of non-NGF ancestors",
			cfg: ancestorConfig{
				numCurrNGFAncestors:    0,
				numCurrNonNGFAncestors: 16,
				numNewNGFAncestors:     0,
			},
			expFull: true,
		},
		{
			name: "current policy is full of a mix of ancestors, but updated list is empty",
			cfg: ancestorConfig{
				numCurrNGFAncestors:    3,
				numCurrNonNGFAncestors: 13,
				numNewNGFAncestors:     0,
			},
			expFull: false,
		},
		{
			name: "current policy is full of NGF ancestors, but updated ancestors is less than that",
			cfg: ancestorConfig{
				numCurrNGFAncestors:    16,
				numCurrNonNGFAncestors: 0,
				numNewNGFAncestors:     5,
			},
			expFull: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			policy := createPolicy(test.cfg)
			full := ngfPolicyAncestorsFull(policy, "nginx-gateway")
			g.Expect(full).To(Equal(test.expFull))
		})
	}
}
