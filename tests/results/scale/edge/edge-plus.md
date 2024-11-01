# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: fed4239ecb35f937b66bba7bd68d6894ca0762b3
- Date: 2024-11-01T00:13:12Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1355000
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 126
- Total Errors: 0
- Average Time: 148ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 384
- Average Time: 122ms
- Event Batch Processing distribution:
	- 500ms: 354
	- 1000ms: 383
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

- Total: 128
- Total Errors: 0
- Average Time: 168ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 451
- Average Time: 116ms
- Event Batch Processing distribution:
	- 500ms: 409
	- 1000ms: 451
	- 5000ms: 451
	- 10000ms: 451
	- 30000ms: 451
	- +Infms: 451

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
- Average Time: 386ms
- Reload distribution:
	- 500ms: 713
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 453ms
- Event Batch Processing distribution:
	- 500ms: 588
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

- Total: 247
- Average Time: 11ms
- Event Batch Processing distribution:
	- 500ms: 247
	- 1000ms: 247
	- 5000ms: 247
	- 10000ms: 247
	- 30000ms: 247
	- +Infms: 247

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
Requests      [total, rate, throughput]         30000, 1000.04, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 924.581µs
Latencies     [min, mean, 50, 90, 95, 99, max]  539.571µs, 717.71µs, 699.684µs, 786.596µs, 819.866µs, 918.033µs, 12.193ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.06, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 776.702µs
Latencies     [min, mean, 50, 90, 95, 99, max]  612.491µs, 803.571µs, 783.584µs, 914.75µs, 962.828µs, 1.065ms, 8.097ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
