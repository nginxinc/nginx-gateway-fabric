---
docs:
---

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
