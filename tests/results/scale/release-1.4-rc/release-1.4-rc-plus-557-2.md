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

- Total: 124
- Total Errors: 0
- Average Time: 163ms
- Reload distribution:
	- 500ms: 124
	- 1000ms: 124
	- 5000ms: 124
	- 10000ms: 124
	- 30000ms: 124
	- +Infms: 124

### Event Batch Processing

- Total: 383
- Average Time: 134ms
- Event Batch Processing distribution:
	- 500ms: 349
	- 1000ms: 380
	- 5000ms: 383
	- 10000ms: 383
	- 30000ms: 383
	- +Infms: 383

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

- Total: 128
- Total Errors: 0
- Average Time: 201ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 451
- Average Time: 132ms
- Event Batch Processing distribution:
	- 500ms: 400
	- 1000ms: 447
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
- Average Time: 930ms
- Reload distribution:
	- 500ms: 425
	- 1000ms: 580
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 1100ms
- Event Batch Processing distribution:
	- 500ms: 357
	- 1000ms: 587
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

- Total: 127
- Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

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
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 1.011ms
Latencies     [min, mean, 50, 90, 95, 99, max]  652.849µs, 1.264ms, 1.009ms, 1.457ms, 2.017ms, 7.354ms, 34.484ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 1.281ms
Latencies     [min, mean, 50, 90, 95, 99, max]  764.722µs, 1.76ms, 1.199ms, 1.879ms, 2.898ms, 13.856ms, 113.63ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
