# Enhancement Proposal-928: Control Plane Dynamic Configuration

- Issue: https://github.com/nginxinc/nginx-kubernetes-gateway/issues/928
- Status: Implementable

## Summary

This proposal contains the design for how to dynamically configure the NGINX Kubernetes Gateway (NKG) control plane.
Through the use of a Custom Resource Definition (CRD), we'll be able to configure fields such as log level or
telemetry at runtime.

## Goals

Define a CRD to dynamically configure mutable options for the NKG control plane. The only initial configurable
option that we will support is log level.

## Non-Goals

This proposal is *not* defining a way to dynamically configure the data plane.

## Introduction

The NKG control plane will evolve to have various user-configurable options. These could include, but are not
limited to, log level, tracing, or metrics. For the best user experience, these options should be able to be
changed at runtime, to avoid having to restart NKG. The first option that we will allow users to configure is the
log level. The easiest and most intuitive way to implement a Kubernetes-native API is through a CRD.

## API, Customer Driven Interfaces, and User Experience

The API would be provided in a CRD. An authorized user would interact with this CRD using `kubectl` to `get`
or `edit` the configuration.

Proposed configuration CRD example:

```yaml
apiVersion: nginx.gateway.k8s.io/v1beta1
kind: NGINXControlConfig
metadata:
    name: nkg-config
    namespace: nginx-gateway
spec:
    logLevel: info
    ...
```

- The CRD would be Namespace-scoped, living in the same Namespace as the controller that it applies to.
- CRD is initialized and created when NKG is deployed
- NKG references the name of this CRD via CLI arg, and only watches this CRD
- If user deletes resource, NKG logs an error and creates an event. Last state is used until CRD is restored.

For discussion with team:

- API group name
- kind name
- default resource name

## Use Cases

The high level use case for dynamically changing settings in the NKG control plane is to allow users to alter
behavior without the need for restarting NKG and experiencing downtime.

For the specific log level use case, users may be experiencing problems with NKG that require more information to
diagnose. These problems could include:

- Not seeing or processing Kubernetes resources that it should be.
- Configuring the data plane incorrectly based on the defined Kubernetes resources.
- Crashes or errors without enough detail.

Being able to dynamically change the log level can allow for a quick way to obtain more information about
the state of the control plane without losing that state due to a required restart.

## Testing

Unit tests can be leveraged for verifying that NKG properly watches and acts on CRD changes. These tests would
be similar in behavior as the current unit tests that verify Gateway API resource processing.

## Security Considerations

We need to ensure that any configurable fields that are exposed to a user are:

- Properly validated. This means that the fields should be the correct type (integer, string, etc.), have appropriate
length, and use regex patterns or enums to prevent any unwanted input.
- Have a valid use case. The more fields we expose, the more attack vectors we create. We should only be exposing
fields that are genuinely useful for a user to change dynamically.

RBAC via the Kubernetes API server will ensure that only authorized users can update the CRD containing NKG control
plane configuration.

## Alternatives

- ConfigMap
A ConfigMap is another type of resource that a user can provide configuration options within, however it lacks the
benefits of a CRD, specifically built-in schema validation, versioning, and conversion webhooks.

- Custom API server
The NKG control plane could implement its own custom API server. However the overhead of implementing this, which
would include auth, validation, endpoints, and so on, would not be worth it due to the fact that the Kubernetes
API server already does all of these things for us.

## References

- [Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
