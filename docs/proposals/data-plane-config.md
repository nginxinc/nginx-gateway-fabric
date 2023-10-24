# Enhancement Proposal-929: Data Plane Dynamic Configuration

- Issue: https://github.com/nginxinc/nginx-kubernetes-gateway/issues/929
- Status: Provisional

## Summary

This proposal is intended to contain the design for how to dynamically configure the data plane for the NGINX Kubernetes
Gateway project. Similar to control plane configuration, we should be able to leverage a custom resource definition to
define data plane configuration.

## Goals

Define a CRD to dynamically configure various settings for the NKG data plane.

## Non-Goals

 - This proposal is not defining every setting that needs to be present in the configuration.
 - This proposal is not for any configuration related to control plane.
