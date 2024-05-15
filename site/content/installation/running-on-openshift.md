---
title: "Deploying NGINX Gateway Fabric in an Openshift environment"
# Change draft status to false to publish doc
draft: false
description: "Learn how to run NGINX Gateway Fabric in an Openshift environment."
weight: 400
toc: true
tags: [ "docs" ]
docs: "DOCS-0000"
---

{{< custom-styles >}}

## Prerequisites

To complete the steps in this guide, you first need to install the following tools for Kubernetes management and development:

- [kubectl](https://kubernetes.io/docs/tasks/tools/): A command-line interface for Kubernetes that allows you to manage and inspect cluster resources, and control containerized applications.
- Access to an Openshift environment with cluster administrative permissions.


## Create SCC Object

In order to deploy NGINX Gateway Fabric instances into Openshift environments, a new SCC object is required to be created
on the cluster which will be used to bind the specific required capabilities to the NGINX Gateway Fabric service account.
To do so for NGF deployments, please run the following command (assuming you are logged in with administrator access to the cluster):

`kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.2.0/deploy/manifests/scc.yaml`

## Deploy NGINX Gateway Fabric

Now that you've created the new SCC, the next step is to install NGINX Gateway Fabric.

We currently only support manual installation with Helm on Openshift:

- [Installation with Helm]({{< relref "installation/installing-ngf/helm.md" >}})

Please follow the instructions in the referenced setup, until you get to running a `helm install` command.
When you do, please add the following flag `--set onOpenshift=true` to whichever `helm install` command you are
using. This will give NGF the correct RBAC permissions to bind to the SCC.

## Getting Started with NGINX Gateway Fabric

Learn how to use NGINX Gateway Fabric by exploring the tutorials in the [examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.2.0/examples) directory. The guides provide practical instructions and scenarios to help you use NGINX Gateway Fabric effectively.

## References

If you have any additional questions specific to the permissions granted in the SCC, feel free to check out
our explanation in our [Openshift Permissions guide]({{< relref "reference/openshift-permissions.md" >}})
