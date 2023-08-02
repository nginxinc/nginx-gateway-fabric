# Enhancement Proposal-554: Control Plane Dynamic Configuration

- Issue: https://github.com/nginxinc/nginx-kubernetes-gateway/issues/554
- Status: Provisional

## Summary

This proposal contains the design for how to dynamically configure the NGINX Kubernetes Gateway (NKG) control plane.
Through the use of a Custom Resource Definition (CRD), we'll be able to configure fields such as log level or
telemetry at runtime.

## Goals

Define a CRD to dynamically configure mutable options for the NKG control plane. The only initial configurable
option that we will support is log level.

## Non-Goals

This proposal is *not* defining a way to dynamically configure the data plane.
