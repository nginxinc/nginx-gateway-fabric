# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
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

- Total: 125
- Total Errors: 0
- Average Time: 147ms
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
	- 500ms: 352
	- 1000ms: 382
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
- Average Time: 167ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 450
- Average Time: 117ms
- Event Batch Processing distribution:
	- 500ms: 408
	- 1000ms: 447
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
- Average Time: 365ms
- Reload distribution:
	- 500ms: 777
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1007
- Average Time: 427ms
- Event Batch Processing distribution:
	- 500ms: 638
	- 1000ms: 1007
	- 5000ms: 1007
	- 10000ms: 1007
	- 30000ms: 1007
	- +Infms: 1007

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

- Total: 211
- Average Time: 9ms
- Event Batch Processing distribution:
	- 500ms: 211
	- 1000ms: 211
	- 5000ms: 211
	- 10000ms: 211
	- 30000ms: 211
	- +Infms: 211

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
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 705.698µs
Latencies     [min, mean, 50, 90, 95, 99, max]  516.719µs, 718.295µs, 696.616µs, 816.483µs, 864.936µs, 991.987µs, 12.915ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             30s, 29.999s, 836.809µs
Latencies     [min, mean, 50, 90, 95, 99, max]  587.011µs, 806.739µs, 782.169µs, 934.845µs, 1.009ms, 1.151ms, 10.413ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
