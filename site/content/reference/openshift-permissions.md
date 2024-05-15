---
title: "Openshift Permissions"
date: 2024-05-14T15:08:41-07:00
# Change draft status to false to publish doc
draft: false
# Description
# Add a short description (150 chars) for the doc. Include keywords for SEO.
# The description text appears in search results and at the top of the doc.
description: "Learn about the permissions given to the NGINX Gateway Fabric Pod in an Openshift environment."
# Assign weights in increments of 100
weight: 100
toc: true
tags: [ "docs" ]
# Create a new entry in the Jira DOCS Catalog and add the ticket ID (DOCS-<number>) below
docs: "DOCS-000"
---

## Overview

To deploy NGINX Gateway Fabric on an Openshift environment, additional permissions are granted to the Pod which are defined
in a SecurityContextConstraints object. This document attempts to describe the permissions given to the Pod and
why they were given.


## Specification

{{< bootstrap-table "table table-bordered table-striped table-responsive" >}}
| Name                       | Value                                                  | Explanation                                                                                                                                                                                                                                                                                                                                                                              |
| -------------------------- | ------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| _allowPrivilegedContainer_ | _false_                                                | None of the containers in the NGINX Gateway Fabric Pod need to be run as privileged.                                                                                                                 |
| _allowHostDirVolumePlugin_ | _false_                                                | None of the containers in the NGINX Gateway Fabric Pod need to use the HostDir volume plugin.                                                                                                        |
| _allowHostIPC_             | _false_                                                | None of the containers in the NGINX Gateway Fabric Pod need host ipc.                                                                                                                                |
| _allowHostNetwork_         | _false_                                                | The NGINX Gateway Fabric Pod does not require the use of HostNetwork.                                                                                                                                |
| _allowHostPorts_           | _false_                                                | None of the containers in the NGINX Gateway Fabric Pod need to use host ports.                                                                                                                       |
| _readOnlyRootFilesystem_   | _true_                                                 | None of the containers in the NGINX Gateway Fabric Pod require a non-read only root file system.                                                                                                     |
| _runAsUser_                | _MustRunAsRange: 101-102_                              | The NGINX Gateway Fabric Pod sets the user ID for the nginx-gateway and nginx containers to 102 and 101 respectively.                                                                                |
| _fsGroup_                  | _MusRunAs: 1001-1001_                                  | The NGINX Gateway Fabric Pod sets the nginx container's fsGroup to 1001.                                                                                                                             |
| _supplementalGroups_       | _MustRunAs: 1001-1001_                                 | Since the nginx container's fsGroup is set to 1001, all processes of the container are also part of the supplementary group ID 1001.                                                                 |
| _seLinuxContext_           | _MustRunAs_                                            | By default, the Kubernetes container runtime assigns the SELinux label to all files on all Pod volumes. Since we don't change any of the labels, we enforce that the labels must be of type SELinux. |
| _volumes_                  | _emptyDir, secret_                                     | TODO                                                                                                                                                                                                     |
| _users_                    | _'system:serviceaccount:*:ngf-nginx-gateway-fabric'_   | This binds the SecurityContextConstraints object to the ServiceAccount associated with the NGINX Gateway Fabric Pod, which gives these permissions to the Pod.                                       |
| _allowedCapabilities_      | _NET_BIND_SERVICE, KILL_                               | TODO                                                                                                                                                                                                     |
| _requiredDropCapabilities_ | _ALL_                                                  | To ensure that the NGINX Gateway Fabric Pod is run with the least amount of capabilities necessary, we drop all the capabilities and only add what's needed.                                         |
{{% /bootstrap-table %}}
