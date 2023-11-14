# How to Install

Add repo:

```text
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
```

Install:

```text
helm install my-opentelemetry-collector open-telemetry/opentelemetry-collector -f values.yaml
```

Uninstall:

```text
helm uninstall my-opentelemetry-collector
```
