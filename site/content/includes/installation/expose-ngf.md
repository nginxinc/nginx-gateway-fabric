---
docs:
---

Gain access to NGINX Gateway Fabric by creating either a **NodePort** service or a **LoadBalancer** service in the same namespace as the controller. The service name is specified in the `--service` argument of the controller.

{{<important>}}
The service manifests configure NGINX Gateway Fabric on ports `80` and `443`, affecting any Gateway [Listeners](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.Listener) on these ports. To use different ports, update the manifests. NGINX Gateway Fabric requires a configured [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/#gateway) resource with a valid listener to listen on any ports.
{{</important>}}

NGINX Gateway Fabric uses this service to update the **Addresses** field in the **Gateway Status** resource. A **LoadBalancer** service sets this field to the IP address and/or hostname. Without a service, the Pod IP address is used.

This gateway is associated with the NGINX Gateway Fabric through the **gatewayClassName** field. The default installation of NGINX Gateway Fabric creates a **GatewayClass** with the name **nginx**. NGINX Gateway Fabric will only configure gateways with a **gatewayClassName** of **nginx** unless you change the name via the `--gatewayclass` [command-line flag](/docs/cli-help.md#static-mode).
