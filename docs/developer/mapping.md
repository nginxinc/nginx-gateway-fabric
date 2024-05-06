# Mapping Gateway API Resources to NGINX Contexts

This document shows how NGINX Gateway Fabric maps Gateway API resources to NGINX contexts. Below are the three types of possible mappings (simplified for brevity):

**A. Distinct Hostnames**

![map-a](/docs/images/mapping/mapping-a.png)

When each HTTPRoute attached to a Gateway has a distinct set of hostnames, NGINX Gateway Fabric will generate a server block for each hostname. All the locations contained in the server block belong to a single HTTPRoute.
In this scenario, each server block is "owned" by a single HTTPRoute.

> Note: This mapping only applies to "Exact" path matches. Some "PathPrefix" path matches will generate an additional location block to handle prefix matching.

**B. Same Hostname**

![map-b](/docs/images/mapping/mapping-b.png)

When HTTPRoutes attached to a Gateway specify the same hostname, NGINX Gateway Fabric generates a single server block containing the locations from all HTTPRoutes that specify the server's hostname.
In this scenario, the HTTPRoutes share "ownership" of the server block.

**C. Internal Redirect**

![map-c](/docs/images/mapping/mapping-c.png)

When HTTPRoutes attached to a Gateway specify the same hostname _and_ path, NGINX Gateway Fabric will generate a single server block and a single (external) location for that path. In addition, it will generate one named location block -- only used for internal requests -- per HTTPRoute match rule.
The external location will offload the routing decision to an NGINX JavaScript (NJS) function that will route the request to the appropriate named location.
In this scenario, the HTTPRoutes share "ownership" of the server block and the external location block.
