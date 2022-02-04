package state

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestUpsertHTTPRoute(t *testing.T) {
	conf := NewFakeConfiguration()

	// Name is set so that the test failure output is easier to comprehend
	hr1 := &v1alpha2.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "route1"}}
	hr2 := &v1alpha2.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "route2"}}

	result := conf.GetArgOfUpsertHTTPRoute()
	if result != nil {
		t.Errorf("GetArgOfUpsertHTTPRoute() returned %v but expected nil", result)
	}

	conf.UpsertHTTPRoute(hr1)

	result = conf.GetArgOfUpsertHTTPRoute()
	if result != hr1 {
		t.Errorf("GetArgOfUpsertHTTPRoute() returned %v but expected %v", result, hr1)
	}

	conf.UpsertHTTPRoute(hr2)

	result = conf.GetArgOfUpsertHTTPRoute()
	if result != hr2 {
		t.Errorf("GetArgOfUpsertHTTPRoute() returned %v but expected %v", result, hr2)
	}
}

func TestDeleteHTTPRoute(t *testing.T) {
	conf := NewFakeConfiguration()

	emtpy := types.NamespacedName{}
	nsname1 := types.NamespacedName{Namespace: "test", Name: "route1"}
	nsname2 := types.NamespacedName{Namespace: "test", Name: "route2"}

	result := conf.GetArgOfDeleteHTTPRoute()
	if result != emtpy {
		t.Errorf("GetArgOfDeleteHTTPRoute() returned %v but expected %v", result, emtpy)
	}

	conf.DeleteHTTPRoute(nsname1)

	result = conf.GetArgOfDeleteHTTPRoute()
	if result != nsname1 {
		t.Errorf("GetArgOfUpsertHTTPRoute() returned %v but expected %v", result, nsname1)
	}

	conf.DeleteHTTPRoute(nsname2)

	result = conf.GetArgOfDeleteHTTPRoute()
	if result != nsname2 {
		t.Errorf("GetArgOfDeleteHTTPRoute() returned %v but expected %v", result, nsname2)
	}
}
