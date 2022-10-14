package reconciler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Getter

// Getter gets a resource from the k8s API.
// It allows us to mock the client.Reader.Get method.
type Getter interface {
	// Get is from client.Reader.
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
}
