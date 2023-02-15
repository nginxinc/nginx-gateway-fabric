/*
Package graph translates the cluster state (Gateway API and Kubernetes resources) into a graph-like representation,
for which:
- Resources are validated. For example, if a Gateway listener is invalid, the graph representation will reflect that.
- Connections between resources are found. For example, if an HTTPRoute attaches to a Gateway, the graph representation
reflects that.
- Any validation or connections-related errors are captured.

Those three points make it easier to generate intermediate data plane configuration and statuses for resources.

The package includes the types to represent the graph and the functions to convert resources into their graph
representation.
*/
package graph
