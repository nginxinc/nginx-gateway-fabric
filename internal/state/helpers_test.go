package state

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestGetNamespacedName(t *testing.T) {
	obj := &v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "hr-1"}}

	expected := types.NamespacedName{Namespace: "test", Name: "hr-1"}

	result := getNamespacedName(obj)
	if result != expected {
		t.Errorf("getNamespacedName() returned %#v but expected %#v", result, expected)
	}
}
