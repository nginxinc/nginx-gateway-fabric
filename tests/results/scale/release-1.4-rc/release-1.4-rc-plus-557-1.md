# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 0f46e3972fab436735d46c051d822f47ab0944e6
- Date: 2024-08-19T15:22:33Z
- Dirty: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1008000
- vCPUs per node: 2
- RAM per node: 4019160Ki
- Max pods per node: 110
- Zone: us-central1-c
- Instance Type: e2-medium

## Test TestScale_Listeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 162ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 385
- Average Time: 132ms
- Event Batch Processing distribution:
	- 500ms: 348
	- 1000ms: 384
	- 5000ms: 385
	- 10000ms: 385
	- 30000ms: 385
	- +Infms: 385

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
- Average Time: 200ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 450
- Average Time: 131ms
- Event Batch Processing distribution:
	- 500ms: 405
	- 1000ms: 447
	- 5000ms: 450
	- 10000ms: 450
	- 30000ms: 450
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
- Average Time: 829ms
- Reload distribution:
	- 500ms: 465
	- 1000ms: 617
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 959ms
- Event Batch Processing distribution:
	- 500ms: 393
	- 1000ms: 624
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
- Average Time: 130ms
- Reload distribution:
	- 500ms: 3
	- 1000ms: 3
	- 5000ms: 3
	- 10000ms: 3
	- 30000ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 208
- Average Time: 44ms
- Event Batch Processing distribution:
	- 500ms: 208
	- 1000ms: 208
	- 5000ms: 208
	- 10000ms: 208
	- 30000ms: 208
	- +Infms: 208

### Errors

- NGF errors: 1
- NGF container restarts: 0
- NGINX errors: 2
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_UpstreamServers) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPMatches

```text
Requests      [total, rate, throughput]         30000, 1000.03, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 1.14ms
Latencies     [min, mean, 50, 90, 95, 99, max]  616.322µs, 1.25ms, 994.898µs, 1.4ms, 1.876ms, 7.545ms, 38.215ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.98
Duration      [total, attack, wait]             30.001s, 29.999s, 1.265ms
Latencies     [min, mean, 50, 90, 95, 99, max]  729.995µs, 2.026ms, 1.213ms, 2.034ms, 3.498ms, 21.007ms, 92.315ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
