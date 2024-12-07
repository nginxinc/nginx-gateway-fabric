# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 929413c15af7bee3adb32e103c9d1513a693da16
- Date: 2024-11-28T12:52:45Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853964Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 126
- Total Errors: 0
- Average Time: 288ms
- Reload distribution:
	- 500.0ms: 126
	- 1000.0ms: 126
	- 5000.0ms: 126
	- 10000.0ms: 126
	- 30000.0ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 385
- Average Time: 172ms
- Event Batch Processing distribution:
	- 500.0ms: 328
	- 1000.0ms: 380
	- 5000.0ms: 385
	- 10000.0ms: 385
	- 30000.0ms: 385
	- +Infms: 385

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

- Total: 128
- Total Errors: 0
- Average Time: 369ms
- Reload distribution:
	- 500.0ms: 99
	- 1000.0ms: 128
	- 5000.0ms: 128
	- 10000.0ms: 128
	- 30000.0ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 450
- Average Time: 171ms
- Event Batch Processing distribution:
	- 500.0ms: 378
	- 1000.0ms: 435
	- 5000.0ms: 450
	- 10000.0ms: 450
	- 30000.0ms: 450
	- +Infms: 450

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
- Average Time: 2672ms
- Reload distribution:
	- 500.0ms: 78
	- 1000.0ms: 179
	- 5000.0ms: 942
	- 10000.0ms: 1001
	- 30000.0ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 2745ms
- Event Batch Processing distribution:
	- 500.0ms: 76
	- 1000.0ms: 171
	- 5000.0ms: 928
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

- Total: 93
- Total Errors: 0
- Average Time: 151ms
- Reload distribution:
	- 500.0ms: 93
	- 1000.0ms: 93
	- 5000.0ms: 93
	- 10000.0ms: 93
	- 30000.0ms: 93
	- +Infms: 93

### Event Batch Processing

- Total: 96
- Average Time: 148ms
- Event Batch Processing distribution:
	- 500.0ms: 96
	- 1000.0ms: 96
	- 5000.0ms: 96
	- 10000.0ms: 96
	- 30000.0ms: 96
	- +Infms: 96

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
Requests      [total, rate, throughput]         29998, 999.93, 999.88
Duration      [total, attack, wait]             30.002s, 30s, 1.525ms
Latencies     [min, mean, 50, 90, 95, 99, max]  555.723µs, 768.075µs, 744.529µs, 872.73µs, 920.635µs, 1.072ms, 12.214ms
Bytes In      [total, mean]                     4859676, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29998  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 786.134µs
Latencies     [min, mean, 50, 90, 95, 99, max]  625.22µs, 858.246µs, 838.767µs, 984.712µs, 1.04ms, 1.162ms, 8.396ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
