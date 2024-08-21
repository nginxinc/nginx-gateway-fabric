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
- Average Time: 163ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 385
- Average Time: 131ms
- Event Batch Processing distribution:
	- 500ms: 351
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

- Total: 125
- Total Errors: 0
- Average Time: 203ms
- Reload distribution:
	- 500ms: 125
	- 1000ms: 125
	- 5000ms: 125
	- 10000ms: 125
	- 30000ms: 125
	- +Infms: 125

### Event Batch Processing

- Total: 446
- Average Time: 130ms
- Event Batch Processing distribution:
	- 500ms: 400
	- 1000ms: 446
	- 5000ms: 446
	- 10000ms: 446
	- 30000ms: 446
	- +Infms: 446

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
- Average Time: 924ms
- Reload distribution:
	- 500ms: 420
	- 1000ms: 577
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 1092ms
- Event Batch Processing distribution:
	- 500ms: 368
	- 1000ms: 584
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

- Total: 134
- Average Time: 30ms
- Event Batch Processing distribution:
	- 500ms: 134
	- 1000ms: 134
	- 5000ms: 134
	- 10000ms: 134
	- 30000ms: 134
	- +Infms: 134

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
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 843.231µs
Latencies     [min, mean, 50, 90, 95, 99, max]  637.486µs, 1.944ms, 985.234µs, 1.469ms, 2.418ms, 17.299ms, 178.442ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.97
Duration      [total, attack, wait]             30.001s, 29.999s, 1.373ms
Latencies     [min, mean, 50, 90, 95, 99, max]  777.081µs, 1.707ms, 1.196ms, 1.925ms, 3.132ms, 14.21ms, 59.211ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
