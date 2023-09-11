package statusfakes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeClient struct {
	// currAttempts is how many times the FakeClient has had its Update or Get method called.
	currAttempts int
	// totalFailedAttempts is how many times the FakeClient will return an error when its Update
	// or Get methods are called before returning nil.
	totalFailedAttempts int
}

// fakeSubResourceClient is created when FakeClient.Status() is called.
// Contains a pointer to FakeClient, so it can update its fields.
type fakeSubResourceClient struct {
	client *FakeClient
}

func NewFakeClient(totalFailedAttempts int) *FakeClient {
	return &FakeClient{0, totalFailedAttempts}
}

func (c *FakeClient) Status() client.SubResourceWriter {
	return c.SubResource("")
}

func (c *FakeClient) SubResource(_ string) client.SubResourceClient {
	return &fakeSubResourceClient{client: c}
}

// Update will return an error sw.totalFailedAttempts times when called by the same Client,
// afterward it will return nil. This will let us test if the status updater retries correctly.
func (sw *fakeSubResourceClient) Update(_ context.Context, _ client.Object, _ ...client.SubResourceUpdateOption) error {
	if sw.client.currAttempts < sw.client.totalFailedAttempts {
		sw.client.currAttempts++
		return fmt.Errorf("client update status failed")
	}
	return nil
}

// Get will return an error c.totalFailedAttempts times when called by the same Client,
// afterward it will return nil. This will let us test if the status updater retries correctly.
func (c *FakeClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	if c.currAttempts < c.totalFailedAttempts {
		c.currAttempts++
		return fmt.Errorf("client get resource failed")
	}
	return nil
}

// Below functions are not used, implemented so fake_client implements Client.

func (sw *fakeSubResourceClient) Get(_ context.Context, _, _ client.Object, _ ...client.SubResourceGetOption) error {
	return nil
}

func (c *FakeClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return nil
}

func (c *FakeClient) Watch(_ context.Context, _ client.ObjectList, _ ...client.ListOption) (watch.Interface, error) {
	return nil, nil
}

func (c *FakeClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return nil
}

func (c *FakeClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return nil
}

func (c *FakeClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
	return nil
}

func (c *FakeClient) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return nil
}

func (c *FakeClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) error {
	return nil
}

func (c *FakeClient) Scheme() *runtime.Scheme {
	return nil
}

func (c *FakeClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (c *FakeClient) GroupVersionKindFor(_ runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (c *FakeClient) IsObjectNamespaced(_ runtime.Object) (bool, error) {
	return false, nil
}

func (sw *fakeSubResourceClient) Create(
	_ context.Context,
	_ client.Object,
	_ client.Object,
	_ ...client.SubResourceCreateOption,
) error {
	return nil
}

func (sw *fakeSubResourceClient) Patch(
	_ context.Context,
	_ client.Object,
	_ client.Patch,
	_ ...client.SubResourcePatchOption,
) error {
	return nil
}
