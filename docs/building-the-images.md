# Building the Images

## Prerequisites

Before you can build the NGINX Kubernetes Gateway and NGINX images, make sure you have the following software
installed on your machine:

- [git](https://git-scm.com/)
- [GNU Make](https://www.gnu.org/software/software.html)
- [Docker](https://www.docker.com/) v18.09+
- [Go](https://go.dev/doc/install) v1.20

## Steps

1. Clone the repo and change into the `nginx-kubernetes-gateway` directory:

   ```shell
   git clone https://github.com/nginxinc/nginx-kubernetes-gateway.git --branch v0.6.0
   cd nginx-kubernetes-gateway
   ```

1. Build the images:
   - To build both the NGINX Kubernetes Gateway and NGINX images:

      ```makefile
      make PREFIX=myregistry.example.com/nginx-kubernetes-gateway build-images
      ```

   - To build just the NGINX Kubernetes Gateway image:

     ```makefile
     make PREFIX=myregistry.example.com/nginx-kubernetes-gateway build-nkg-image
     ```

   - To build just the NGINX image:

     ```makefile
     make PREFIX=myregistry.example.com/nginx-kubernetes-gateway build-nginx-image
     ```

   Set the `PREFIX` variable to the name of the registry you'd like to push the image to. By default, the images will be
   named `nginx-kubernetes-gateway:0.6.0` and `nginx-kubernetes-gateway/nginx:0.6.0`.

1. Push the images to your container registry:

   ```shell
   docker push myregistry.example.com/nginx-kubernetes-gateway:0.6.0
   docker push myregistry.example.com/nginx-kubernetes-gateway/nginx:0.6.0
   ```

   Make sure to substitute `myregistry.example.com/nginx-kubernetes-gateway` with your registry.
