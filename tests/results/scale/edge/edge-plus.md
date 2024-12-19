# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853984Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 235ms
- Reload distribution:
	- 500.0ms: 128
	- 1000.0ms: 128
	- 5000.0ms: 128
	- 10000.0ms: 128
	- 30000.0ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 387
- Average Time: 155ms
- Event Batch Processing distribution:
	- 500.0ms: 341
	- 1000.0ms: 384
	- 5000.0ms: 387
	- 10000.0ms: 387
	- 30000.0ms: 387
	- +Infms: 387

### Errors

- NGF errors: 2
- NGF container restarts: 0
- NGINX errors: 12
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_Listeners) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPSListeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 264ms
- Reload distribution:
	- 500.0ms: 128
	- 1000.0ms: 128
	- 5000.0ms: 128
	- 10000.0ms: 128
	- 30000.0ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 451
- Average Time: 146ms
- Event Batch Processing distribution:
	- 500.0ms: 388
	- 1000.0ms: 451
	- 5000.0ms: 451
	- 10000.0ms: 451
	- 30000.0ms: 451
	- +Infms: 451

### Errors

- NGF errors: 0
- NGF container restarts: 0
- NGINX errors: 17
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_HTTPSListeners) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPRoutes

### Reloads

- Total: 1001
- Total Errors: 0
- Average Time: 1575ms
- Reload distribution:
	- 500.0ms: 131
	- 1000.0ms: 307
	- 5000.0ms: 1001
	- 10000.0ms: 1001
	- 30000.0ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 1695ms
- Event Batch Processing distribution:
	- 500.0ms: 118
	- 1000.0ms: 280
	- 5000.0ms: 1008
	- 10000.0ms: 1008
	- 30000.0ms: 1008
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
- Average Time: 151ms
- Reload distribution:
	- 500.0ms: 3
	- 1000.0ms: 3
	- 5000.0ms: 3
	- 10000.0ms: 3
	- 30000.0ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 84
- Average Time: 177ms
- Event Batch Processing distribution:
	- 500.0ms: 84
	- 1000.0ms: 84
	- 5000.0ms: 84
	- 10000.0ms: 84
	- 30000.0ms: 84
	- +Infms: 84

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
Duration      [total, attack, wait]             30s, 30s, 458.338µs
Latencies     [min, mean, 50, 90, 95, 99, max]  367.922µs, 471.815µs, 459.996µs, 520.013µs, 548.43µs, 654.523µs, 10.646ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.03, 999.92
Duration      [total, attack, wait]             30.002s, 29.999s, 3.368ms
Latencies     [min, mean, 50, 90, 95, 99, max]  435.427µs, 2.982ms, 3.17ms, 3.855ms, 4.253ms, 7.523ms, 32.46ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
