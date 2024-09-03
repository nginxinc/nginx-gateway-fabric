# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855004Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 125
- Total Errors: 0
- Average Time: 145ms
- Reload distribution:
	- 500ms: 125
	- 1000ms: 125
	- 5000ms: 125
	- 10000ms: 125
	- 30000ms: 125
	- +Infms: 125

### Event Batch Processing

- Total: 384
- Average Time: 124ms
- Event Batch Processing distribution:
	- 500ms: 345
	- 1000ms: 384
	- 5000ms: 384
	- 10000ms: 384
	- 30000ms: 384
	- +Infms: 384

### Errors

- NGF errors: 3
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
- Average Time: 164ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 449
- Average Time: 120ms
- Event Batch Processing distribution:
	- 500ms: 407
	- 1000ms: 446
	- 5000ms: 449
	- 10000ms: 449
	- 30000ms: 449
	- +Infms: 449

### Errors

- NGF errors: 2
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
	- 500ms: 842
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 406ms
- Event Batch Processing distribution:
	- 500ms: 698
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

- Total: 137
- Total Errors: 0
- Average Time: 126ms
- Reload distribution:
	- 500ms: 137
	- 1000ms: 137
	- 5000ms: 137
	- 10000ms: 137
	- 30000ms: 137
	- +Infms: 137

### Event Batch Processing

- Total: 140
- Average Time: 125ms
- Event Batch Processing distribution:
	- 500ms: 140
	- 1000ms: 140
	- 5000ms: 140
	- 10000ms: 140
	- 30000ms: 140
	- +Infms: 140

### Errors

- NGF errors: 2
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_UpstreamServers) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPMatches

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 666.43µs
Latencies     [min, mean, 50, 90, 95, 99, max]  520.85µs, 724.76µs, 686.355µs, 776.205µs, 811.753µs, 923.775µs, 21.111ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 739.661µs
Latencies     [min, mean, 50, 90, 95, 99, max]  569.875µs, 773.833µs, 759.109µs, 866.32µs, 912.478µs, 1.021ms, 10.63ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
