# Enhancement Proposal-2035: Advanced NGINX Extensions

- Issue: https://github.com/nginxinc/nginx-gateway-fabric/issues/2035
- Status: Provisional

## Summary

NGINX Gateway Fabric (NGF) [exposes](/site/content/overview/gateway-api-compatibility.md) NGINX features via Gateway API
resources (like HTTPRoute) and [NGINX extensions](nginx-extensions.md) (like ClientSettingPolicy). Combined, they
expose a subset of the most common NGINX configuration. As we implement more Gateway API resources and NGINX extensions,
the subset will grow. However, it will take time. Additionally, because the number of NGINX configuration directives
and parameters is huge, not all of them will be supported that way. As a result, users are not able to implement certain
NGINX use cases. To allow them to implement those use cases, we need to bring a new extension mechanism to NGF.

## Goals

- Allow users to insert NGINX configuration not supported via Gateway API resources or NGINX extensions.
- Allow users to customize supported NGINX configuration (for example, add a parameter to an NGINX directive).
- Support configuration from modules not loaded in NGINX by default or third-party modules.
- Most of the configuration complexity should fall onto the cluster operator persona, not the application developer.
- Provide security controls to prevent application developers from injecting arbitrary NGINX configuration.
- Ensure adequate configuration validation to prevent NGINX outages due to invalid configuration.
- Advanced NGINX extensions can be used without source code modification.

## Non-Goals

- Support configuration other than NGINX directives. For example, njs configuration files or TLS certificates.
- Reimplement already supported features through the new extension mechanism.
