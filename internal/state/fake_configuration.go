package state

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// FakeConfiguration is for testing purposes.
type FakeConfiguration struct {
	upsertHTTPRoute *v1alpha2.HTTPRoute
	deleteHTTPRoute types.NamespacedName

	lock sync.RWMutex
}

// NewFakeConfiguration creates a new FakeConfiguration.
func NewFakeConfiguration() *FakeConfiguration {
	return &FakeConfiguration{}
}

// UpsertHTTPRoute implements a fake UpsertHTTPRoute of Configuration.
func (c *FakeConfiguration) UpsertHTTPRoute(httpRoute *v1alpha2.HTTPRoute) ([]Change, []StatusUpdate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.upsertHTTPRoute = httpRoute
	return nil, nil
}

// GetArgOfUpsertHTTPRoute returns the HTTPRoute passed in the latest invocation of the UpsertHTTPRoute method.
func (c *FakeConfiguration) GetArgOfUpsertHTTPRoute() *v1alpha2.HTTPRoute {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.upsertHTTPRoute
}

// DeleteHTTPRoute implements a fake DeleteHTTPRoute of Configuration.
func (c *FakeConfiguration) DeleteHTTPRoute(nsname types.NamespacedName) ([]Change, []StatusUpdate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.deleteHTTPRoute = nsname
	return nil, nil
}

// GetArgOfDeleteHTTPRoute returns the NamespacedName passed in the latest invocation of the DeleteHTTPRoute method.
func (c *FakeConfiguration) GetArgOfDeleteHTTPRoute() types.NamespacedName {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.deleteHTTPRoute
}
