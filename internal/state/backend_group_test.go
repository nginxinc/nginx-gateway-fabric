package state_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestBackendGroup_GroupName(t *testing.T) {
	bg := state.BackendGroup{
		Source:  types.NamespacedName{Namespace: "test", Name: "hr"},
		RuleIdx: 20,
	}
	expected := "test__hr_rule20"
	result := bg.GroupName()
	if result != expected {
		t.Errorf("BackendGroup.GroupName() mismatch; expected %s, got %s", expected, result)
	}
}
