---
title: "Expose NGINX Gateway Fabric"
weight: 300
docs: "DOCS-1427"
---

There are two options for accessing NGINX Gateway Fabric depending on the type of LoadBalancer service you chose during installation:

- If the LoadBalancer type is `NodePort`, Kubernetes will randomly allocate two ports on every node of the cluster.
  To access the NGINX Gateway Fabric, use an IP address of any node of the cluster along with the two allocated ports.

  {{<tip>}} Read more about the type NodePort in the [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport). {{</tip>}}

- If the LoadBalancer type is `LoadBalancer`:

  - For GCP or Azure, Kubernetes will allocate a cloud load balancer for load balancing the NGINX Gateway Fabric pods.
    Use the public IP of the load balancer to access NGINX Gateway Fabric.
  - For AWS, Kubernetes will allocate a Network Load Balancer (NLB) in TCP mode with the PROXY protocol enabled to pass
    the client's information (the IP address and the port).

  Use the public IP of the load balancer to access NGINX Gateway Fabric. To get the public IP which is reported in the `EXTERNAL-IP` column:

  - For GCP or Azure, run:

    ```shell
    kubectl get svc nginx-gateway -n nginx-gateway
    ```

  - In AWS, the NLB (Network Load Balancer) DNS (directory name system) name will be reported by Kubernetes instead of a public IP. To get the DNS name, run:

    ```shell
    kubectl get svc nginx-gateway -n nginx-gateway
    ```

    {{< note >}} We recommend using the NLB DNS whenever possible, but for testing purposes, you can resolve the DNS name to get the IP address of the load balancer:

  ```shell
  nslookup <dns-name>
  ```

    {{< /note >}}

  {{<tip>}} Learn more about type LoadBalancer in the [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/service/#type-loadbalancer).

  For AWS, additional options regarding an allocated load balancer are available, such as its type and SSL
  termination. Read the [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/service/#type-loadbalancer) to learn more.
  {{</tip>}}

{{<important>}}By default Helm and manifests configure NGINX Gateway Fabric on ports `80` and `443`, affecting any gateway [listeners](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.Listener) on these ports. To use different ports, update the configuration. NGINX Gateway Fabric requires a configured [gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener to listen on any ports.{{</important>}}

NGINX Gateway Fabric uses the created service to update the **Addresses** field in the **Gateway Status** resource. Using a **LoadBalancer** service sets this field to the IP address and/or hostname of that service. Without a service, the pod IP address is used.

This gateway is associated with the NGINX Gateway Fabric through the **gatewayClassName** field. The default installation of NGINX Gateway Fabric creates a **GatewayClass** with the name **nginx**. NGINX Gateway Fabric will only configure gateways with a **gatewayClassName** of **nginx** unless you change the name via the `--gatewayclass` [command-line flag](/docs/cli-help.md#static-mode).
