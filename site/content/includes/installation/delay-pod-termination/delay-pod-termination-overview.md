---
docs: "DOCS-1441"
---

To avoid client service interruptions when upgrading NGINX Gateway Fabric, you can configure [`PreStop` hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/) to delay terminating the NGINX Gateway Fabric pod, allowing the pod to complete certain actions before shutting down. This ensures a smooth upgrade without any downtime, also known as a zero downtime upgrade.

For an in-depth explanation of how Kubernetes handles pod termination, see the [Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) topic on their official website.

{{<note>}}Keep in mind that NGINX won't shut down while WebSocket or other long-lived connections are open. NGINX will only stop when these connections are closed by the client or the backend. If these connections stay open during an upgrade, Kubernetes might need to shut down NGINX forcefully. This sudden shutdown could interrupt service for clients.{{</note>}}
