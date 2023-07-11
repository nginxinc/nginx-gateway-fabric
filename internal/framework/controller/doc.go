/*
Package controller is responsible for creating and registering controllers for
sigs.k8s.io/controller-runtime/pkg/manager.Manager.

A controller is responsible for watching for updates to the resource of a desired type and propagating those updates
as events through the event channel.

The reconciliation part of a controller -- reacting on a resource change -- is implemented by the Reconciler type,
which in turn implements sigs.k8s.io/controller-runtime/pkg/reconcile.Reconciler.
*/
package controller
