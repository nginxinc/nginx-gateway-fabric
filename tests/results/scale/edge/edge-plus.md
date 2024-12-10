# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 929413c15af7bee3adb32e103c9d1513a693da16
- Date: 2024-11-28T12:52:45Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 233ms
- Reload distribution:
	- 500.0ms: 128
	- 1000.0ms: 128
	- 5000.0ms: 128
	- 10000.0ms: 128
	- 30000.0ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 387
- Average Time: 165ms
- Event Batch Processing distribution:
	- 500.0ms: 333
	- 1000.0ms: 385
	- 5000.0ms: 387
	- 10000.0ms: 387
	- 30000.0ms: 387
	- +Infms: 387

### Errors

- NGF errors: 1
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
- Average Time: 255ms
- Reload distribution:
	- 500.0ms: 128
	- 1000.0ms: 128
	- 5000.0ms: 128
	- 10000.0ms: 128
	- 30000.0ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 451
- Average Time: 156ms
- Event Batch Processing distribution:
	- 500.0ms: 385
	- 1000.0ms: 446
	- 5000.0ms: 451
	- 10000.0ms: 451
	- 30000.0ms: 451
	- +Infms: 451

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
- Average Time: 1481ms
- Reload distribution:
	- 500.0ms: 142
	- 1000.0ms: 341
	- 5000.0ms: 1001
	- 10000.0ms: 1001
	- 30000.0ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 1604ms
- Event Batch Processing distribution:
	- 500.0ms: 122
	- 1000.0ms: 297
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

- Total: 2
- Total Errors: 0
- Average Time: 152ms
- Reload distribution:
	- 500.0ms: 2
	- 1000.0ms: 2
	- 5000.0ms: 2
	- 10000.0ms: 2
	- 30000.0ms: 2
	- +Infms: 2

### Event Batch Processing

- Total: 47
- Average Time: 383ms
- Event Batch Processing distribution:
	- 500.0ms: 39
	- 1000.0ms: 45
	- 5000.0ms: 47
	- 10000.0ms: 47
	- 30000.0ms: 47
	- +Infms: 47

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
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 608.314µs
Latencies     [min, mean, 50, 90, 95, 99, max]  519.534µs, 725.239µs, 682.818µs, 788.076µs, 832.943µs, 1.021ms, 20.567ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 877.596µs
Latencies     [min, mean, 50, 90, 95, 99, max]  585.521µs, 772.61µs, 753.696µs, 876.012µs, 928.327µs, 1.057ms, 10.962ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
