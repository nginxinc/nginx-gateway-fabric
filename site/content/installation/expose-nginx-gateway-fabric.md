---
title: "Expose NGINX Gateway Fabric"
description: ""
weight: 300
toc: true
docs: "DOCS-1427"
---

{{<custom-styles>}}

## Overview

Gain access to NGINX Gateway Fabric by creating either a **NodePort** service or a **LoadBalancer** service in the same namespace as the controller. The service name is specified in the `--service` argument of the controller.

{{<important>}}The service manifests configure NGINX Gateway Fabric on ports `80` and `443`, affecting any gateway [listeners](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.Listener) on these ports. To use different ports, update the manifests. NGINX Gateway Fabric requires a configured [gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener to listen on any ports.{{</important>}}

NGINX Gateway Fabric uses the created service to update the **Addresses** field in the **Gateway Status** resource. Using a **LoadBalancer** service sets this field to the IP address and/or hostname of that service. Without a service, the pod IP address is used.

This gateway is associated with the NGINX Gateway Fabric through the **gatewayClassName** field. The default installation of NGINX Gateway Fabric creates a **GatewayClass** with the name **nginx**. NGINX Gateway Fabric will only configure gateways with a **gatewayClassName** of **nginx** unless you change the name via the `--gatewayclass` [command-line flag](/docs/cli-help.md#static-mode).

## Create a NodePort service

To create a **NodePort** service run the following command:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.3.0/deploy/manifests/service/nodeport.yaml
```

A **NodePort** service allocates a port on every cluster node. Access NGINX Gateway Fabric using any node's IP address and the allocated port.

## Create a LoadBalancer Service

To create a **LoadBalancer** service, use the appropriate manifest for your cloud provider:

### GCP (Google Cloud Platform) and Azure

1. Run the following command:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.3.0/deploy/manifests/service/loadbalancer.yaml
   ```

2. Lookup the public IP of the load balancer, which is reported in the `EXTERNAL-IP` column in the output of the following command:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

3. Use the public IP of the load balancer to access NGINX Gateway Fabric.

### AWS (Amazon Web Services)

1. Run the following command:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.3.0/deploy/manifests/service/loadbalancer-aws-nlb.yaml
   ```

2. In AWS, the NLB (Network Load Balancer) DNS (directory name system) name will be reported by Kubernetes instead of a public IP in the `EXTERNAL-IP` column. To get the DNS name, run:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

   {{< note >}} We recommend using the NLB DNS whenever possible, but for testing purposes, you can resolve the DNS name to get the IP address of the load balancer:

   ```shell
   nslookup <dns-name>
   ```

   {{< /note >}}
