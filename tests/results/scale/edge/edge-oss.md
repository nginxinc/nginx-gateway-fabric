# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: cba529e9b53cc011fc6f5e539164eefe779ab7c9
- Date: 2024-09-03T18:57:20Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 145ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 386
- Average Time: 116ms
- Event Batch Processing distribution:
	- 500ms: 353
	- 1000ms: 386
	- 5000ms: 386
	- 10000ms: 386
	- 30000ms: 386
	- +Infms: 386

### Errors

- NGF errors: 2
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_Listeners) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPSListeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 165ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 450
- Average Time: 116ms
- Event Batch Processing distribution:
	- 500ms: 405
	- 1000ms: 450
	- 5000ms: 450
	- 10000ms: 450
	- 30000ms: 450
	- +Infms: 450

### Errors

- NGF errors: 1
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
- Average Time: 347ms
- Reload distribution:
	- 500ms: 835
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 404ms
- Event Batch Processing distribution:
	- 500ms: 719
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

- Total: 144
- Total Errors: 0
- Average Time: 127ms
- Reload distribution:
	- 500ms: 144
	- 1000ms: 144
	- 5000ms: 144
	- 10000ms: 144
	- 30000ms: 144
	- +Infms: 144

### Event Batch Processing

- Total: 147
- Average Time: 126ms
- Event Batch Processing distribution:
	- 500ms: 147
	- 1000ms: 147
	- 5000ms: 147
	- 10000ms: 147
	- 30000ms: 147
	- +Infms: 147

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
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 665.633µs
Latencies     [min, mean, 50, 90, 95, 99, max]  537.484µs, 768.096µs, 734.32µs, 849.425µs, 893.692µs, 1.02ms, 20.405ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 966.342µs
Latencies     [min, mean, 50, 90, 95, 99, max]  596.705µs, 861.342µs, 843.414µs, 991.162µs, 1.049ms, 1.168ms, 12.784ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
