# Gateway Evaluation

The SIG-NETWORK group has begun a mission to improve traffic control to and from kubernetes clusters. The Gateway API aims to improve service networking by formalizing user personas and developing a decoupled and composable collection of resources. The intent is to create an extensible and expressive API that lends itself naturally to role-based access control methods.

## Personas

Gateway API focuses on three main personas: Infrastructure Provider (Infra), Cluster Operator (Admin), and Application Developer (AppDev); for more details see [Roles and personas](https://gateway-api.sigs.k8s.io/concepts/security-model/#roles-and-personas).

The Infra role focuses on provider tier assets: compute (e.g. virtual machines, base-metal server/node hardware), storage (e.g. external arrays, SANs), and networking (e.g. VPNs, subnet isolation, routing). NGINX does not provide Infra services or a platform product as it pertains to this role. For the purposes of this evaluation we intend to focus on the Admin and AppDev personas. The Gateway evaluation will require (and in some cases assume) a correct infrastructure and network topology is in place; cloud network routing, SDN control, DNS, and L4 load balancing are out of scope.

## Resources

The Gateway evaluation intends to solve for a subset of L7 traffic management use-cases. The Gateway evaluation will not replace or supplant current NIC use-cases and will not work to achieve feature parity. The Gateway evaluation will not support Ingress v1 APIs or set of NIC CRDs (e.g. VirtualServer and VirtualServerRoute).

Three main API kinds will be supported for chosen Core features with Optional and Extended features to be considered later.

### GatewayClass
Defines a set of gateways with a common configuration and behavior.
### Gateway
Requests a point where traffic can be translated to Services within the cluster.
### Routes
Describe how traffic coming via the Gateway maps to the Services.

The Gateway evaluation will employ each in the following manner:
- GatewayClass: The Admin controls the resource lifecycle and is responsible for creating, updating, and deleting a GatewayClass for NGINX. The GatewayClass is used by the Admin to provide global configuration; by way of ParameterRefs, to the subordinate Gateway resources.
- Gateway: The Admin controls the resource lifecycle and is responsible for creating, updating, and deleting a Gateway for NGINX. The Gateway is a reprensentation and description of an NGINX Gateway deployment - an NGINX Gateway deployment is made up of a set of NGINX dataplane and controlplane pairs.
- Routes: The Gateway API describes multiple typed Routes: HTTP, TLS, TCP, and UDP. The AppDev controls the resource lifecycle and is responsible for creating, updating, and delete a Route. The Gateway evaluation will only support a subset of HTTPRoute features and not the entire Core set.

## Goals
- Vet a minimal feature set of Core resources for their applicability and conformance to NGINX configuration models.
- Support L7 HTTPRoute resources; host routing and path routing only.
- Support GatewayClass, Gateway, and HTTPRoute Status subobject requirements.
- Develop expertise in Gateway API and refine a compatible, extensible, and flexible architecture.

## Design
NGINX Gateway is a decoupled (two containers) controlplane and dataplane.

The dataplane is NGINX OSS run as a container within a Kubernetes Pod with `.spec.shareProcessNamespace` set to true.

The controlplane is a manager of a set of Go controllers built on `sigs.k8s.io/controller-runtime`. The controlplane will provide a layered separation between an API SDK, runtime controllers, and implementation. The separation isn't strictly necessary and may be abandonned due to time constraints, but a decoupled SDK and implementation will provide flexibility and forward compatibility options.

The initial controllers required are GatewayClass, Gateway, HTTPRoute, and ParametersRef (implementation detail and yet to be determined API GVK/GVR).

Component diagram:

```
---------------    ------------------    -----------------------    ----------------------
| Controllers | -> | SDK Reconciler | -> | SDK Impl. Reconcile | -> | Graph / AST / Conf |
---------------    ------------------    -----------------------    ----------------------
                     ^                     ^
                     |                     |
                   -------               -------------
                   | SDK | <- Register - | SDK Impl. |
                   -------               -------------
```

NGINX Gateway will be deployed inside the Kubernetes cluster using Deployments, DaemonSets, or unmanaged Pods and must be a pair of containers - the dataplane container and the controlplane container. (TBD: leader election, controller-runtime provides facilities but requirements need defining.)

Expected deployment:
```
                Cluster Boundary
                |                        --------------
                |                        | API Server |
                |                        --------------
                |                          |
                |                          V
                |                        --------------------
                |                        | ---------------- |
                |                        | | ControlPlane | |         ------------
                |                        | ---------------- |       - | Upstream |
                |                        |   |              |      /  ------------
                |                        |   V              |     /
------------    |    ----------------    | -------------    |    /    ------------
| Ext.Load | -> | -> | Svc Endpoint | -> | | DataPlane | -> | ------- | Upstream |
| Balancer |    |    ----------------    | -------------    |    \
------------    |                        --------------------     \   ------------
                |                                                  \- | Upstream |
                |                                                     ------------
```

### Prerequisites
- Infra persona has enabled external access to the cluster. For example, configured an external network device which has connectivity to the cluster Node underlay network.
- Admin has knowledge of the cluster external DNS Hostname and listen Port.

### Deployment Requirements
- Admin provisions GatewayClass referencing Gateway object name.
- Admin provisions Gateway using name referenced from GatewayClass.
- Admin provisions Gateway using hostname and port: `.spec.listeners[].{ hostname, port, protocol }` where protocol MUST be "HTTP".
- Admin provisions cluster access to NGINX dataplanes via Service abstractions: Service `.spec.type=LoadBalancer` or `.spec.type=NodePort`.
- Admin provisions NGINX Gateway deployment via workload abstractions, Deployment or DaemonSet.

### Operation

#### Startup
At process startup, the NGINX Gateway must discover the "state of the world".

NGINX Gateway will,
- initialize with a minimal configuration; reading an environment variable or command-line argument referencing the GatewayClass name.
- watch for and wait for its GatewayClass.
  - log its state and subsequent state change (GatewayClass found or not found, GatewayClass discovered and process proceeding).
- watch for ParametersRef: `.spec.parametersRef`. If not found, continue processing with default runtime configuration.
- determine Gateway resource from GatewayClass: `.spec.controllerName` (TBD: is this true?)
- watch and wait for its Gateway as referenced by the discovered gateway controller.
- bootstrap internal data models with `.spec.listeners[].hostname` for route matching.
- instantiate controller and control loop for HTTPRoute resources.
- rebuild current declarative state using stored HTTPRoute resources.
- enter runtime loop.

(TBD: analyze and remark where Status updates occur)

#### Runtime
The NGINX Gateway is primarily a set of control loops watching for changes in the declared configuration of the Gateway API.

Process runtime configuration updates may be delivered by updates to GatewayClass, Gateway, and ParametersRef. NGINX Gateway will watch these resources for updates, reconciling their changes, and resolving or rejecting based on calculated conflicts.

Dataplane traffic decisions and configuration are primarily conveyed via HTTPRoute updates. NGINX Gateway will establish an HTTPRoute controller reconciling configuration updates, and resolving or rejecting based on calculated conflicts or validation errors. Route updates are translated into an intermediate representation; data structure specifics are yet to determined. Properties desired: sparse, directed, acyclic, relational.

Once validated for conflicts and errors, the intermediate representation is exported to the dataplane as NGINX configuration directives.

Loop cycle:
- Event notification: Upsert, Delete
- Reconcilation:
  - Delete
    - Remove data structure nodes
    - Validate, commit or rollback
    - Write config
  - Upsert
    - Add or Update data structure nodes
    - Validate, commit or rollback
    - Write config
