---
title: "Expose NGINX Gateway Fabric"
date: 2023-11-16T16:37:42-08:00
# Change draft status to false to publish doc.
draft: false
# Description
# Add a short description (150 chars) for the doc. Include keywords for SEO. 
# The description text appears in search results and at the top of the doc.
description: ""
# Assign weights in increments of 100
weight: 300
toc: true
tags: [ "docs" ]
# Create a new entry in the Jira DOCS Catalog and add the ticket ID (DOCS-<number>) below
docs: "DOCS-000"
# Taxonomies
# These are pre-populated with all available terms for your convenience.
# Remove all terms that do not apply.
categories: ["installation", "platform management", "load balancing", "api management", "service mesh", "security", "analytics"]
doctypes: ["task"]
journeys: ["researching", "getting started", "using", "renewing", "self service"]
personas: ["devops", "netops", "secops", "support"]
versions: []
authors: []

---

### Expose NGINX Gateway Fabric


Gain access to NGINX Gateway Fabric by creating either a **NodePort** service or a **LoadBalancer** service in the same namespace as the controller. The service name is specified in the `--service` argument of the controller.

{{<important>}}
The service manifests configure NGINX Gateway Fabric on ports `80` and `443`, affecting any Gateway [Listeners](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener) on these ports. To use different ports, update the manifests. NGINX Gateway Fabric requires a configured [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener to listen on any ports.
{{</important>}}

NGINX Gateway Fabric uses this service to update the **Addresses** field in the **Gateway Status** resource. A **LoadBalancer** service sets this field to the IP address and/or hostname. Without a service, the Pod IP address is used.

This gateway is associated with the NGINX Gateway Fabric through the **gatewayClassName** field. The default installation of NGINX Gateway Fabric creates a **GatewayClass** with the name **nginx**. NGINX Gateway Fabric will only configure gateways with a **gatewayClassName** of **nginx** unless you change the name via the `--gatewayclass` [command-line flag](/docs/cli-help.md#static-mode).

#### Create a NodePort service

To create a **NodePort** service:

```shell
kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/nodeport.yaml
```

A **NodePort** service allocates a port on every cluster node. Access NGINX Gateway Fabric using any node's IP address and the allocated port.

#### Create a LoadBalancer Service

To create a **LoadBalancer** service, use the appropriate manifest for your cloud provider:

- For GCP (Google Cloud Platform) or Azure:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/loadbalancer.yaml
   ```

  Lookup the public IP of the load balancer, which is reported in the `EXTERNAL-IP` column in the output of the following command:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  Use the public IP of the load balancer to access NGINX Gateway Fabric.

- For AWS (Amazon Web Services):

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/nginxinc/nginx-gateway-fabric/v1.0.0/deploy/manifests/service/loadbalancer-aws-nlb.yaml
   ```

  In AWS, the NLB (Network Load Balancer) DNS (directory name system) name will be reported by Kubernetes instead of a public IP in the `EXTERNAL-IP` column. To get the DNS name, run:

   ```shell
   kubectl get svc nginx-gateway -n nginx-gateway
   ```

  Generally, use the NLB DNS name, but for testing purposes, you can resolve the DNS name to get the IP address of the load balancer:

   ```shell
   nslookup <dns-name>
   ```
