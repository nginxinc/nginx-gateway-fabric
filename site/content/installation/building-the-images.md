---
title: "Building NGINX Gateway Fabric and NGINX Images"
weight: 300
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Overview

While most users will install NGINX Gateway Fabric [with Helm]({{< relref "/installation/installing-ngf/helm.md" >}}) or [Kubernetes manifests]({{< relref "/installation/installing-ngf/manifests.md" >}}), manually building the [NGINX Gateway Fabric and NGINX images]({{< relref "/overview/gateway-architecture.md#the-nginx-gateway-fabric-pod" >}}) can be helpful for testing and development purposes. Follow the steps in this document to build the NGINX Gateway Fabric and NGINX images.

## Prerequisites

Before you can build the NGINX Gateway Fabric and NGINX images, make sure you have the following software
installed on your machine:

- [git](https://git-scm.com/)
- [GNU Make](https://www.gnu.org/software/software.html)
- [Docker](https://www.docker.com/) v18.09+
- [Go](https://go.dev/doc/install) v1.20

## Steps

1. Clone the repo and change into the `nginx-gateway-fabric` directory:

   ```shell
   git clone https://github.com/nginxinc/nginx-gateway-fabric.git
   cd nginx-gateway-fabric
   ```

1. Build the images:
   - To build both the NGINX Gateway Fabric and NGINX images:

      ```makefile
      make PREFIX=myregistry.example.com/nginx-gateway-fabric build-images
      ```

   - To build just the NGINX Gateway Fabric image:

     ```makefile
     make PREFIX=myregistry.example.com/nginx-gateway-fabric build-ngf-image
     ```

   - To build just the NGINX image:

     ```makefile
     make PREFIX=myregistry.example.com/nginx-gateway-fabric build-nginx-image
     ```

   Set the `PREFIX` variable to the name of the registry you'd like to push the image to. By default, the images will be
   named `nginx-gateway-fabric:edge` and `nginx-gateway-fabric/nginx:edge`.

1. Push the images to your container registry:

   ```shell
   docker push myregistry.example.com/nginx-gateway-fabric:edge
   docker push myregistry.example.com/nginx-gateway-fabric/nginx:edge
   ```

   Make sure to substitute `myregistry.example.com/nginx-gateway-fabric` with your registry.
