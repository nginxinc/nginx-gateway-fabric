# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 0f46e3972fab436735d46c051d822f47ab0944e6
- Date: 2024-08-19T15:22:33Z
- Dirty: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-central1-c
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 148ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 385
- Average Time: 124ms
- Event Batch Processing distribution:
	- 500ms: 354
	- 1000ms: 383
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

- Total: 126
- Total Errors: 0
- Average Time: 171ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 449
- Average Time: 117ms
- Event Batch Processing distribution:
	- 500ms: 406
	- 1000ms: 448
	- 5000ms: 449
	- 10000ms: 449
	- 30000ms: 449
	- +Infms: 449

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
- Average Time: 364ms
- Reload distribution:
	- 500ms: 767
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 418ms
- Event Batch Processing distribution:
	- 500ms: 658
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
- Average Time: 127ms
- Reload distribution:
	- 500ms: 3
	- 1000ms: 3
	- 5000ms: 3
	- 10000ms: 3
	- 30000ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 262
- Average Time: 14ms
- Event Batch Processing distribution:
	- 500ms: 262
	- 1000ms: 262
	- 5000ms: 262
	- 10000ms: 262
	- 30000ms: 262
	- +Infms: 262

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
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 694.128µs
Latencies     [min, mean, 50, 90, 95, 99, max]  524.19µs, 720.737µs, 705.561µs, 821.627µs, 868.751µs, 992.126µs, 10.214ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 805.125µs
Latencies     [min, mean, 50, 90, 95, 99, max]  598.092µs, 833.481µs, 807.105µs, 970.861µs, 1.044ms, 1.247ms, 13.085ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
