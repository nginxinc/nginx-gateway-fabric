---
title: "Push an NGINX Plus image to a private registry"
weight: 200
doctypes: ["install"]
toc: true
docs: "DOCS-1433"
---

## Overview

This document describes how to pull a NGINX Plus image for NGINX Gateway Fabric from the official F5 Docker registry and upload it to your private registry

## Before you begin

Before you start, you'll need these installed on your machine:

- [Docker v18.09 or higher](https://docs.docker.com/engine/release-notes/18.09/).
- The certificate (**nginx-repo.crt**) and key (**nginx-repo.key**) for a Connectivity Stack for Kubernetes subscription, obtainable from [MyF5l](https://my.f5.com) An NGINX Plus certificate and key will not work.

## Configuring Docker for the F5 Container Registry

To configure Docker to communicate with the F5 Container Registry, first create a folder containing your certificate and key files:

```shell
mkdir -p /etc/docker/certs.d/private-registry.nginx.com
cp <path-to-your-nginx-repo.crt> /etc/docker/certs.d/private-registry.nginx.com/client.cert
cp <path-to-your-nginx-repo.key> /etc/docker/certs.d/private-registry.nginx.com/client.key
```

If you are not using a Linux operating system, read the [Docker for Windows](https://docs.docker.com/desktop/faqs/windowsfaqs/#how-do-i-add-custom-ca-certificates) or [Docker for Mac](https://docs.docker.com/desktop/faqs/macfaqs/#add-custom-ca-certificates-server-side) instructions. For more details on Docker Engine security, you can refer to the [Docker Engine Security documentation](https://docs.docker.com/engine/security/).


## Pulling the image

Once configured, you can now pull images from `private-registry.nginx.com`. To find your desired image, read the [Technical Specifications](https://github.com/nginxinc/nginx-gateway-fabric#technical-specifications).

Run this command step to pull an image, replacing `<version-tag>` with the specific version you need, such as `1.4.0`.


  ```shell
  docker pull private-registry.nginx.com/nginx-gateway-fabric/nginx-plus:1.4.0
  ```

You can use the Docker registry API to list available image tags using your client certificate and key. The `jq` command is used to format the JSON output for easier reading.

```shell
curl https://private-registry.nginx.com/nginx-gateway-fabric/nginx-plus/tags/list --key <path-to-client.key> --cert <path-to-client.cert> | jq
```

```json
{
  "name": "nginx-gateway-fabric/nginx-plus",
  "tags": [
    "edge",
    "nightly"
  ]
}
```


Once you have pulled an image, you can tag it and push it to a private registry.

1. Log into your private registry:

   ```shell
   docker login <my-docker-registry>
   ```

1. Tag the image, replacing `<my-docker-registry>` with your registry's path and `<version-tag>` with the version you're using:


    ```shell
    docker tag private-registry.nginx.com/nginx-gateway-fabric/nginx-plus:<version-tag> <my-docker-registry>/nginx-gateway-fabric/nginx-plus:<version-tag>
    docker push <my-docker-registry>/nginx-gateway-fabric/nginx-plus:<version-tag>
    ```


## Troubleshooting

If you encounter issues while following this guide, here are solutions to common problems:

- **Certificate errors**:
  - *Likely cause*: Incorrect certificate or key location, or using an NGINX Plus certificate.
  - *Solution*: Check you have the correct NGINX Gateway Fabric certificate and key, their files are named correctly, and they are in the correct directory.

- **Docker version compatibility**
  - *Likely cause*: Outdated Docker version.
  - *Solution*: Make sure you're running [Docker v18.09 or higher](https://docs.docker.com/engine/release-notes/18.09/), and upgrade if necessary.

- **Can't pull the image**
  - *Likely cause*: Mismatched image name or tag.
  - *Solution*: Compare the image name and tag to the [Technical Specifications table](https://github.com/nginxinc/nginx-gateway-fabric?tab=readme-ov-file#technical-specifications).

- **Failed to push to private registry**
  - *Likely cause*: Not logged into your private registry or incorrect image tagging.
  - *Solution*: Verify your login status and correct the image tag before pushing. Read the [Docker documentation](https://docs.docker.com/docker-hub/repos/) for more guidance.


## Alternative installation options

There are alternative ways to get an NGINX Plus image for NGINX Gateway Fabric:

- [Install by pulling a docker image]({{<relref "jwt-token-docker-secret.md#pulling-an-image-for-local-use">}}).
- [Build the Gateway Fabric image]({{<relref "installation/ngf-images/building-the-images.md">}}) using the source code from the GitHub repository and your NGINX Plus subscription certificate and key.
