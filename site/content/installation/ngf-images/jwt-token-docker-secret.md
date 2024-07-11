---
title: "Getting the NGINX Plus image using JWT"
description: "This document describes how to use a JWT token to get an NGINX Plus image for NGINX Gateway Fabric from the F5 Docker registry."
weight: 100
doctypes: ["install"]
toc: true
docs: "DOCS-1432"
---

## Overview

Follow the steps in this document to pull the NGINX Plus image for NGINX Gateway Fabric from the F5 Docker registry into your Kubernetes cluster using your JWT token. To list available image tags using the Docker registry API, you will also need to download its certificate and key from [MyF5](https://my.f5.com).

{{<important>}}An NGINX Plus subscription will not work with these instructions. For NGINX Gateway Fabric, you must have an Connectivity Stack for Kubernetes subscription.{{</important>}}

## Before you begin

You will need the following items from [MyF5](https://my.f5.com) for these instructions:

1. A JWT Access Token for NGINX Gateway Fabric from an active Connectivity Stack for Kubernetes subscription (Per instance).
1. The certificate (**nginx-repo.crt**) and key (**nginx-repo.key**) for each NGINX Gateway Fabric instance.

## Get the Credentials

1. Log into the [MyF5 Portal](https://my.f5.com/), navigate to your subscription details, and download the required certificate, key and JWT files.

## Using the JWT token in a Docker Config Secret

1. Create a Kubernetes `docker-registry` secret type on the cluster, using the contents of the JWT token as the username and `none` for password (as the password is not used).  The name of the docker server is `private-registry.nginx.com`.

    ```shell
    kubectl create secret docker-registry nginx-plus-registry-secret --docker-server=private-registry.nginx.com --docker-username=<JWT Token> --docker-password=none [-n nginx-gateway]
    ```

   It is important that the `--docker-username=<JWT Token>` contains the contents of the token and is not pointing to the token itself. When you copy the contents of the JWT token, ensure there are no additional characters such as extra whitespaces. This can invalidate the token, causing 401 errors when trying to authenticate to the registry.

1. Inspect and verify the details of the created secret by running:

    ```shell
    kubectl get secret nginx-plus-registry-secret --output=yaml
    ```

{{< include "installation/jwt-password-note.md" >}}

## Install NGINX Gateway Fabric

Please refer to [Installing NGINX Gateway Fabric]({{< relref "installation/installing-ngf" >}})


## Pulling an image for local use

To pull an image for local use, use this command:

```shell
docker login private-registry.nginx.com --username=<output_of_jwt_token> --password=none
```

Replace the contents of `<output_of_jwt_token>` with the contents of the JWT token itself.
Once you have successfully pulled the image, you can tag it as needed, then push it to a different container registry.


## Alternative installation options

There are alternative ways to get an NGINX Plus image for NGINX Gateway Fabric:

- [Build the Gateway Fabric image]({{<relref "installation/ngf-images/building-the-images">}}) describes how to use the source code with an NGINX Plus subscription certificate and key to build an image.
