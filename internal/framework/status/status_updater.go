package status

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . StatusUpdater

// StatusUpdater updates a resource from the k8s API.
// It allows us to mock the client.Reader.Status.Update method.
//
// nolint:revive
type StatusUpdater interface {
	// Update is from client.StatusClient.SubResourceWriter.
	Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error
}
