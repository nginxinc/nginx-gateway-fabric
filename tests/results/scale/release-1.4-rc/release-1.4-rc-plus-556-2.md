# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: c0b0e63b22d5e68e8fe029ce224a877544d9e999
- Date: 2024-08-21T22:16:10Z
- Dirty: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-central1-c
- Instance Type: n2d-standard-16

## Test TestScale_UpstreamServers

### Reloads

- Total: 3
- Total Errors: 0
- Average Time: 126ms
- Reload distribution:
	- 500ms: 3
	- 1000ms: 3
	- 5000ms: 3
	- 10000ms: 3
	- 30000ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 206
- Average Time: 9ms
- Event Batch Processing distribution:
	- 500ms: 206
	- 1000ms: 206
	- 5000ms: 206
	- 10000ms: 206
	- 30000ms: 206
	- +Infms: 206

### Errors

- NGF errors: 0
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_UpstreamServers) for more details.
The logs are attached only if there are errors.
