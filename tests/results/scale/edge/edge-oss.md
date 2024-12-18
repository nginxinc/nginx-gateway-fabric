# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853980Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 222ms
- Reload distribution:
	- 500.0ms: 128
	- 1000.0ms: 128
	- 5000.0ms: 128
	- 10000.0ms: 128
	- 30000.0ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 386
- Average Time: 149ms
- Event Batch Processing distribution:
	- 500.0ms: 341
	- 1000.0ms: 385
	- 5000.0ms: 386
	- 10000.0ms: 386
	- 30000.0ms: 386
	- +Infms: 386

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

- Total: 127
- Total Errors: 0
- Average Time: 242ms
- Reload distribution:
	- 500.0ms: 127
	- 1000.0ms: 127
	- 5000.0ms: 127
	- 10000.0ms: 127
	- 30000.0ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 449
- Average Time: 140ms
- Event Batch Processing distribution:
	- 500.0ms: 397
	- 1000.0ms: 444
	- 5000.0ms: 449
	- 10000.0ms: 449
	- 30000.0ms: 449
	- +Infms: 449

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
- Average Time: 1564ms
- Reload distribution:
	- 500.0ms: 133
	- 1000.0ms: 307
	- 5000.0ms: 1001
	- 10000.0ms: 1001
	- 30000.0ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 1648ms
- Event Batch Processing distribution:
	- 500.0ms: 127
	- 1000.0ms: 281
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

- Total: 141
- Total Errors: 0
- Average Time: 151ms
- Reload distribution:
	- 500.0ms: 141
	- 1000.0ms: 141
	- 5000.0ms: 141
	- 10000.0ms: 141
	- 30000.0ms: 141
	- +Infms: 141

### Event Batch Processing

- Total: 144
- Average Time: 149ms
- Event Batch Processing distribution:
	- 500.0ms: 144
	- 1000.0ms: 144
	- 5000.0ms: 144
	- 10000.0ms: 144
	- 30000.0ms: 144
	- +Infms: 144

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
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 548.6µs
Latencies     [min, mean, 50, 90, 95, 99, max]  513.431µs, 686.307µs, 661.593µs, 762.241µs, 810.258µs, 961.421µs, 14.726ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 722.836µs
Latencies     [min, mean, 50, 90, 95, 99, max]  580.249µs, 769.326µs, 746.54µs, 877.525µs, 936.325µs, 1.082ms, 10.123ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
