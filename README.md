[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/nginxinc/nginx-gateway-fabric/badge)](https://api.securityscorecards.dev/projects/github.com/nginxinc/nginx-gateway-fabric)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-gateway-fabric.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-gateway-fabric?ref=badge_shield)

# NGINX Gateway Fabric

NGINX Gateway Fabric is an open-source project that provides an implementation of
the [Gateway API](https://gateway-api.sigs.k8s.io/) using [NGINX](https://nginx.org/) as the data plane. The goal of
this project is to implement the core Gateway APIs -- `Gateway`, `GatewayClass`, `HTTPRoute`, `TCPRoute`, `TLSRoute`,
and `UDPRoute` -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running
on Kubernetes. NGINX Gateway Fabric supports a subset of the Gateway API.

For a list of supported Gateway API resources and features, see
the [Gateway API Compatibility](https://docs.nginx.com/nginx-gateway-fabric/overview/gateway-api-compatibility/) doc.

Learn about our [design principles](/docs/developer/design-principles.md) and [architecture](https://docs.nginx.com/nginx-gateway-fabric/overview/gateway-architecture/).

## Getting Started

1. [Quick Start on a kind cluster](https://docs.nginx.com/nginx-gateway-fabric/installation/running-on-kind/).
2. [Install](https://docs.nginx.com/nginx-gateway-fabric/installation/) NGINX Gateway Fabric.
3. [Build](https://docs.nginx.com/nginx-gateway-fabric/installation/building-the-images/) an NGINX Gateway Fabric container image from source or use a pre-built image
   available
   on [GitHub Container Registry](https://github.com/nginxinc/nginx-gateway-fabric/pkgs/container/nginx-gateway-fabric).
4. Deploy various [examples](examples).
5. Read our [How-to guides](https://docs.nginx.com/nginx-gateway-fabric/how-to/).

You can find the comprehensive NGINX Gateway Fabric user documentation on the [NGINX Documentation](https://docs.nginx.com/nginx-gateway-fabric/) website.

## NGINX Gateway Fabric Releases

We publish NGINX Gateway Fabric releases on GitHub. See
our [releases page](https://github.com/nginxinc/nginx-gateway-fabric/releases).

The latest release is [1.1.0](https://github.com/nginxinc/nginx-gateway-fabric/releases/tag/v1.1.0).

The edge version is useful for experimenting with new features that are not yet published in a release. To use, choose
the *edge* version built from the [latest commit](https://github.com/nginxinc/nginx-gateway-fabric/commits/main)
from the main branch.

The table below summarizes the options regarding the images, manifests, documentation and examples and gives your links
to the correct versions:

| Version        | Description                              | Installation Manifests                                                            | Documentation and Examples                                                                                                                                                     |
|----------------|------------------------------------------|-----------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Latest release | For production use                       | [Manifests](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.1.0/deploy). | [Documentation](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.1.0/site/content). [Examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/v1.1.0/examples). |
| Edge           | For experimental use and latest features | [Manifests](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/deploy).   | [Documentation](https://docs.nginx.com/nginx-gateway-fabric/). [Examples](https://github.com/nginxinc/nginx-gateway-fabric/tree/main/examples).                                |

### Versioning

NGF uses semantic versioning for its releases. For more information, see https://semver.org.

> Major version zero `(0.Y.Z)` is reserved for development, anything MAY change at any time. The public API is not stable.

### Release Planning and Development

The features that will go into the next release are reflected in the
corresponding [milestone](https://github.com/nginxinc/nginx-gateway-fabric/milestones). Refer to
the [Issue Lifecycle](ISSUE_LIFECYCLE.md) document for information on issue creation and assignment to releases.


## Technical Specifications

The following table lists the software versions NGINX Gateway Fabric supports.

| NGINX Gateway Fabric | Gateway API | Kubernetes | NGINX OSS |
|----------------------|-------------|------------|-----------|
| Edge                 | 1.0.0       | 1.23+      | 1.25.3    |
| 1.1.0                | 1.0.0       | 1.23+      | 1.25.3    |
| 1.0.0                | 0.8.1       | 1.23+      | 1.25.2    |
| 0.6.0                | 0.8.0       | 1.23+      | 1.25.2    |
| 0.5.0                | 0.7.1       | 1.21+      | 1.25.x *  |
| 0.4.0                | 0.7.1       | 1.21+      | 1.25.x *  |
| 0.3.0                | 0.6.2       | 1.21+      | 1.23.x *  |
| 0.2.0                | 0.5.1       | 1.21+      | 1.21.x *  |
| 0.1.0                | 0.5.0       | 1.19+      | 1.21.3    |

\*the installation manifests use the minor version of NGINX container image (e.g. 1.25) and the patch version is not
specified. This means that the latest available patch version is used.

## SBOM (Software Bill of Materials)

We generate SBOMs for the binaries and the Docker image.

### Binaries

The SBOMs for the binaries are available in the releases page. The SBOMs are generated
using [syft](https://github.com/anchore/syft) and are available in SPDX format.

### Docker Images

The SBOM for the Docker image is available in
the [GitHub Container](https://github.com/nginxinc/nginx-gateway-fabric/pkgs/container/nginx-gateway-fabric)
repository. The SBOM is generated using [syft](https://github.com/anchore/syft) and stored as an attestation in the
image manifest.

For example to retrieve the SBOM for `linux/amd64` and analyze it using [grype](https://github.com/anchore/grype) you
can run the following command:

```shell
docker buildx imagetools inspect ghcr.io/nginxinc/nginx-gateway-fabric:edge --format '{{ json (index .SBOM "linux/amd64").SPDX }}' | grype
```

## Troubleshooting

For troubleshooting help, see the [Troubleshooting](https://docs.nginx.com/nginx-gateway-fabric/how-to/monitoring/troubleshooting/) document.

## Contacts

We’d like to hear your feedback! If you experience issues with our Gateway Controller, please [open a bug][bug] in
GitHub. If you have any suggestions or enhancement requests, please [open an idea][idea] on GitHub discussions. You can
contact us directly via kubernetes@nginx.com or on the [NGINX Community Slack][slack] in
the `#nginx-gateway-fabric`
channel.

[bug]:https://github.com/nginxinc/nginx-gateway-fabric/issues/new?assignees=&labels=&projects=&template=bug_report.md&title=

[idea]:https://github.com/nginxinc/nginx-gateway-fabric/discussions/categories/ideas

[slack]: https://nginxcommunity.slack.com/channels/nginx-gateway-fabric

## Contributing

Please read our [Contributing guide](CONTRIBUTING.md) if you'd like to contribute to the project.

## Support

NGINX Gateway Fabric is not covered by any support contract.
