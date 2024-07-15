# Results

## Test environment

NGINX Plus: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.4-gke.1043004
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 146ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 387
- Average Time: 128ms
- Event Batch Processing distribution:
	- 500ms: 348
	- 1000ms: 387
	- 5000ms: 387
	- 10000ms: 387
	- 30000ms: 387
	- +Infms: 387

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
- Average Time: 166ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 449
- Average Time: 122ms
- Event Batch Processing distribution:
	- 500ms: 403
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
- Average Time: 373ms
- Reload distribution:
	- 500ms: 743
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1007
- Average Time: 432ms
- Event Batch Processing distribution:
	- 500ms: 638
	- 1000ms: 1006
	- 5000ms: 1007
	- 10000ms: 1007
	- 30000ms: 1007
	- +Infms: 1007

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

- Total: 104
- Average Time: 117ms
- Event Batch Processing distribution:
	- 500ms: 104
	- 1000ms: 104
	- 5000ms: 104
	- 10000ms: 104
	- 30000ms: 104
	- +Infms: 104

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
Duration      [total, attack, wait]             30s, 29.999s, 895.124µs
Latencies     [min, mean, 50, 90, 95, 99, max]  745.651µs, 1.018ms, 991.503µs, 1.133ms, 1.178ms, 1.304ms, 27.128ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.98
Duration      [total, attack, wait]             30.001s, 29.999s, 1.178ms
Latencies     [min, mean, 50, 90, 95, 99, max]  801.928µs, 1.099ms, 1.079ms, 1.24ms, 1.307ms, 1.447ms, 12.938ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
