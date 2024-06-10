---
title: "Enabling Usage Reporting for NGINX Plus"
description: "This page outlines how to enable Usage Reporting for NGINX Gateway Fabric and how to view the usage data through the API."
weight: 1000
toc: true
docs: "DOCS-000"
---

## Overview

Usage Reporting connects to the NGINX Instance Manager and reports the number of Nodes and NGINX Gateway Fabric Pods in the cluster.

To use Usage Reporting, you must have access to [NGINX Instance Manager](https://www.nginx.com/products/nginx-management-suite/instance-manager). Usage Reporting is a requirement of the new Flexible Consumption Program for NGINX Gateway Fabric, used to calculate costs. **This only applies if using NGINX Plus as the data plane.** Usage is reported every 24 hours.

## Requirements

Usage Reporting needs to be configured when deploying NGINX Gateway Fabric.

To enable Usage Reporting, you must have the following:

- NGINX Gateway Fabric 1.2.0 or later
- [NGINX Instance Manager 2.11](https://docs.nginx.com/nginx-management-suite) or later

In addition to the software requirements, you will need:

- Access to an NGINX Instance Manager username and password for basic authentication. You will also need the URL of your NGINX Instance Manager system. The Usage Reporting user account must have access to the `/api/platform/v1/k8s-usage` endpoint.
- Access to the Kubernetes cluster where NGINX Gateway Fabric is deployed, with the ability to deploy a Kubernetes Secret.

## Adding a User Account to NGINX Instance Manager

1. Create a role following the steps in the [Create Roles](https://docs.nginx.com/nginx-management-suite/admin-guides/rbac/create-roles/) section of the NGINX Instance Manager documentation. Select these permissions in step 6 for the role:

- Module: Instance Manager
- Feature: NGINX Plus Usage
- Access: CRUD

1. Create a user account following the steps in the [Create New Users](https://docs.nginx.com/nginx-management-suite/admin-guides/authentication/basic-authentication/#create-users) section of the NGINX Instance Manager documentation. In step 6, assign the user to the role created above. Note that currently only "Basic Auth" authentication is supported for Usage Reporting purposes.

## Enabling Usage Reporting in NGINX Gateway Fabric

### Adding Credentials to a Kubernetes Secret

To make the credentials available to NGINX Gateway Fabric to connect to the NGINX Instance Manager, we need to create a Kubernetes Secret.

1. The username and password should be base64 encoded and stored in the Secret. In the following example, the username is `foo` and the password is `bar`. Run a similar command to generate the base64 encoded strings of your username and password:

   ```shell
   echo -n 'foo' | base64
   # Zm9v
   echo -n 'bar' | base64
   # YmFy
   ```

1. Create the following Secret in your Kubernetes cluster, replacing the username and password with your generated strings. You can rename the Secret if desired, and create it in any Namespace you want.

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
      name: ngf-usage-auth
      namespace: nginx-gateway
   type: kubernetes.io/basic-auth
   data:
      username: Zm9v # base64 representation of 'foo' obtained in step 1
      password: YmFy # base64 representation of 'bar' obtained in step 1
   ```

   If you need to update the basic-auth credentials at any time, update the `username` and `password` fields and apply the changes. NGINX Gateway Fabric will automatically detect the changes and use the new username and password without redeployment.

### Install NGINX Gateway Fabric with Usage Reporting enabled

When installing NGINX Gateway Fabric, a few configuration options need to be specified in order to enable Usage Reporting. You should follow the normal [installation](https://docs.nginx.com/nginx-gateway-fabric/installation/) steps using your preferred method, but ensure you include the following options:

If using Helm, the `nginx.usage` values should be set as necessary:

- `secretName` should be the `namespace/name` of the credentials Secret you created. Using our example, it would be `nginx-gateway/ngf-usage-auth`. This field is required.
- `serverURL` is the base server URL of the NGINX Instance Manager. This field is required.
- `clusterName` is an optional display name in the API for the usage data object.

If using manifests, the following command-line options should be set as necessary on the `nginx-gateway` container:

- `--usage-report-secret` should be the `namespace/name` of the credentials Secret you created. Using our example, it would be `nginx-gateway/ngf-usage-auth`. This field is required.
- `--usage-report-server-url` is the base server URL of the NGINX Instance Manager. This field is required.
- `--usage-report-cluster-name` is an optional display name in the API for the usage data object.

Your NGINX Gateway Fabric Pods should also have one of the following labels:

- `app.kubernetes.io/name=nginx-gateway`
- `app.kubernetes.io/name=nginx-gateway-fabric`

{{< note >}}The default installation of NGINX Gateway Fabric already includes at least one of these labels.{{< /note >}}

## Viewing Usage Data from the NGINX Instance Manager API

NGINX Gateway Fabric sends the number of its instances and nodes in the cluster to NGINX Instance Manager every 24 hours. To view the usage data, query the NGINX Instance Manager API. The usage data is available at the following endpoint (replace `nim.example.com` with your server URL, and set the proper credentials in the `--user` field):

```shell
curl --user "foo:bar" https://nim.example.com/api/platform/v1/k8s-usage
```

```json
{
  "items": [
    {
      "max_node_count": 5,
      "metadata": {
        "createTime": "2023-01-27T09:12:33.001Z",
        "displayName": "my-cluster",
        "monthReturned": "May",
        "uid": "d290f1ee-6c54-4b01-90e6-d701748f0851",
        "updateTime": "2023-01-29T10:12:33.001Z"
      },
      "node_count": 4,
      "pod_details": {
        "current_pod_counts": {
          "dos_count": 0,
          "pod_count": 15,
          "waf_count": 0
        },
        "max_pod_counts": {
         "max_dos_count": 0,
          "max_pod_count": 25,
          "max_waf_count": 0
        }
      }
    },
    {
      "max_node_count": 3,
      "metadata": {
        "createTime": "2023-01-25T09:12:33.001Z",
        "displayName": "my-cluster2",
        "monthReturned": "May",
        "uid": "12tgb8ug-g8ik-bs7h-gj3j-hjitk672946hb",
        "updateTime": "2023-01-26T10:12:33.001Z"
      },
      "node_count": 3,
      "pod_details": {
        "current_pod_counts": {
          "dos_count": 0,
          "pod_count": 5,
          "waf_count": 0
        },
        "max_pod_counts": {
          "max_dos_count": 0,
          "max_pod_count": 15,
          "max_waf_count": 0
        }
      }
    }
  ]
}
```

You can also query the usage data for a specific cluster by specifying the cluster uid in the endpoint, for example:

```shell
curl --user "foo:bar" https://nim.example.com/api/platform/v1/k8s-usage/d290f1ee-6c54-4b01-90e6-d701748f0851
```

```json
{
  "max_node_count": 5,
  "metadata": {
    "createTime": "2023-01-27T09:12:33.001Z",
    "displayName": "my-cluster",
    "monthReturned": "May",
    "uid": "d290f1ee-6c54-4b01-90e6-d701748f0851",
    "updateTime": "2023-01-29T10:12:33.001Z"
  },
  "node_count": 4,
  "pod_details": {
    "current_pod_counts": {
      "dos_count": 0,
      "pod_count": 15,
      "waf_count": 0
    },
    "max_pod_counts": {
      "max_dos_count": 0,
      "max_pod_count": 25,
      "max_waf_count": 0
    }
  }
}
```
