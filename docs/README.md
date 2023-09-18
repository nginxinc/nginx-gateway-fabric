# NGINX Kubernetes Gateway Documentation

This directory contains all of the documentation relating to NGINX Kubernetes Gateway.

## Contents

- [Architecture](architecture.md): An overview of the architecture and design principles of NGINX Kubernetes Gateway.
- [Gateway API Compatibility](gateway-api-compatibility.md): Describes which Gateway API resources NGINX Kubernetes
Gateway supports and the extent of that support.
- [Installation](installation.md): Walkthrough on how to install NGINX Kubernetes Gateway on a generic Kubernetes cluster.
- [Resource Validation](resource-validation.md): Describes how NGINX Kubernetes Gateway validates Gateway API
resources.
- [Control Plane Configuration](control-plane-configuration.md): Describes how to dynamically update the NGINX
Kubernetes Gateway control plane configuration.
- [Building the Images](building-the-images.md): Steps on how to build the NGINX Kubernetes Gateway container images
yourself.
- [Running on Kind](running-on-kind.md): Walkthrough on how to run NGINX Kubernetes Gateway on a `kind` cluster.
- [CLI Help](cli-help.md): Describes the commands available in the `gateway` binary of `nginx-kubernetes-gateway`
container.
- [Monitoring](monitoring.md): Information on monitoring NGINX Kubernetes Gateway using Prometheus metrics.
- [Troubleshooting](troubleshooting.md): Troubleshooting guide for common or known issues.

### Directories

- [Guides](guides): Guides about configuring NGINX Kubernetes Gateway for various use cases.
- [Developer](developer/): Docs for developers of the project. Contains guides relating to processes and workflows.
- [Proposals](proposals/): Enhancement proposals for new features.
