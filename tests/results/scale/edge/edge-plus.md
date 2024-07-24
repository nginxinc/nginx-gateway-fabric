# Results

## Test environment

NGINX Plus: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1038001
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 126
- Total Errors: 0
- Average Time: 146ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 384
- Average Time: 129ms
- Event Batch Processing distribution:
	- 500ms: 347
	- 1000ms: 382
	- 5000ms: 384
	- 10000ms: 384
	- 30000ms: 384
	- +Infms: 384

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

- Total: 126
- Total Errors: 0
- Average Time: 165ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 448
- Average Time: 120ms
- Event Batch Processing distribution:
	- 500ms: 400
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
- Average Time: 344ms
- Reload distribution:
	- 500ms: 841
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 397ms
- Event Batch Processing distribution:
	- 500ms: 726
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
- Average Time: 126ms
- Reload distribution:
	- 500ms: 3
	- 1000ms: 3
	- 5000ms: 3
	- 10000ms: 3
	- 30000ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 43
- Average Time: 159ms
- Event Batch Processing distribution:
	- 500ms: 41
	- 1000ms: 43
	- 5000ms: 43
	- 10000ms: 43
	- 30000ms: 43
	- +Infms: 43

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
Duration      [total, attack, wait]             30s, 29.999s, 730.824µs
Latencies     [min, mean, 50, 90, 95, 99, max]  557.726µs, 742.366µs, 725.104µs, 820.304µs, 857.029µs, 954.105µs, 17.509ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 772.125µs
Latencies     [min, mean, 50, 90, 95, 99, max]  625.466µs, 823.002µs, 799.21µs, 918.366µs, 964.408µs, 1.072ms, 24.548ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
