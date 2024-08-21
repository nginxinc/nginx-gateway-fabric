# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 0f46e3972fab436735d46c051d822f47ab0944e6
- Date: 2024-08-19T15:22:33Z
- Dirty: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-central1-c
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 126
- Total Errors: 0
- Average Time: 149ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 383
- Average Time: 121ms
- Event Batch Processing distribution:
	- 500ms: 350
	- 1000ms: 383
	- 5000ms: 383
	- 10000ms: 383
	- 30000ms: 383
	- +Infms: 383

### Errors

- NGF errors: 0
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_Listeners) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPSListeners

### Reloads

- Total: 125
- Total Errors: 0
- Average Time: 169ms
- Reload distribution:
	- 500ms: 125
	- 1000ms: 125
	- 5000ms: 125
	- 10000ms: 125
	- 30000ms: 125
	- +Infms: 125

### Event Batch Processing

- Total: 448
- Average Time: 118ms
- Event Batch Processing distribution:
	- 500ms: 405
	- 1000ms: 448
	- 5000ms: 448
	- 10000ms: 448
	- 30000ms: 448
	- +Infms: 448

### Errors

- NGF errors: 0
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_HTTPSListeners) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPRoutes

### Reloads

- Total: 1001
- Total Errors: 0
- Average Time: 360ms
- Reload distribution:
	- 500ms: 791
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 415ms
- Event Batch Processing distribution:
	- 500ms: 675
	- 1000ms: 1008
	- 5000ms: 1008
	- 10000ms: 1008
	- 30000ms: 1008
	- +Infms: 1008

### Errors

- NGF errors: 0
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_HTTPRoutes) for more details.
The logs are attached only if there are errors.

## Test TestScale_UpstreamServers

### Reloads

- Total: 3
- Total Errors: 0
- Average Time: 125ms
- Reload distribution:
	- 500ms: 3
	- 1000ms: 3
	- 5000ms: 3
	- 10000ms: 3
	- 30000ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 243
- Average Time: 11ms
- Event Batch Processing distribution:
	- 500ms: 243
	- 1000ms: 243
	- 5000ms: 243
	- 10000ms: 243
	- 30000ms: 243
	- +Infms: 243

### Errors

- NGF errors: 0
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_UpstreamServers) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPMatches

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 777.707µs
Latencies     [min, mean, 50, 90, 95, 99, max]  560.632µs, 834.012µs, 781.794µs, 914.15µs, 965.138µs, 1.208ms, 22.203ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 787.986µs
Latencies     [min, mean, 50, 90, 95, 99, max]  614.744µs, 873.454µs, 855.24µs, 1.001ms, 1.057ms, 1.181ms, 13.113ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
