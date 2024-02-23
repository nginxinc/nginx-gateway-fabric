---
title: "Collecting Product Telemetry"
description: "Learn how NGINX Gateway Fabric collects product telemetry to help make improvements in the product."
weight: 500
toc: true
---

## Overview

In order to understand our users and the features that they value, NGINX Gateway Fabric collects product telemetry by default. This data is used to help determine the priorities of the product to ensure we are delivering the most valuable functionality and features to our users.

Telemetry data is collected once every 24 hours and sent to a service managed by F5. Personally identifiable information (PII) is **not** collected. The list of data points that are collected can be seen below.

**If you would prefer to not have data collected, you can [opt-out](#opt-out) when installing NGINX Gateway Fabric.**

## Collected Data

- **Kubernetes Platform and Version:** the Kubernetes platform and version that NGINX Gateway Fabric is running on.
- **Platform Architecture:** the architecture that the Kubernetes environment is running on.
- **Cluster UID:** the UID of the `kube-system` Namespace in the cluster that NGINX Gateway Fabric is running in.
- **Cluster Node Count:** the number of Nodes in the cluster.
- **Deployment UID:** the UID of the NGINX Gateway Fabric Deployment.
- **Deployment Replica Count:** the count of NGINX Gateway Fabric Pods.
- **Image Build Source:** whether the image was built by Github or locally (values are `gha`, `local`, or `unknown`). The source repository of the images is **not** collected.
- **NGINX Modules:** a list of installed NGINX modules.
- **Deployment Flags:** a list of NGINX Gateway Fabric Deployment flags that are specified by a user. The actual values of non-boolean flags are **not** collected; we only record that they are either `default` or `user-defined`.
- **Count of Relevant Resources:** the total count of relevant resources to NGINX Gateway Fabric. This includes `GatewayClasses`, `Gateways`, `HTTPRoutes`, `Secrets`, `Services`, and `Endpoints`. The data within these resources is **not** collected.

This data is used to identify the following information:

- The flavors of Kubernetes environments that are most popular among our users.
- The number of unique NGINX Gateway Fabric installations.
- The scale of NGINX Gateway Fabric Deployments.
- The scale of Gateway API resources.
- The used features of NGINX Gateway Fabric.

Our goal is to publicly discuss data trends to drive roadmap discussions in our [Community Meeting](https://github.com/nginxinc/nginx-gateway-fabric/discussions/1472).

## Opt Out

To disable the collection of product telemetry, set one of the following options when installing NGINX Gateway Fabric, depending on your installation method:

### Helm

Set the `nginxGateway.productTelemetry.enable=false` flag either in the `values.yaml` file or when installing:

```shell
helm install ... --set nginxGateway.productTelemetry.enable=false
```

### Manifests

Add the `--product-telemetry-disable` flag to the `nginx-gateway` container in your Deployment manifest.
