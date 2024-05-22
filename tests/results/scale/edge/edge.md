# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 12
- k8s version: v1.28.8-gke.1095000
- vCPUs per node: 16
- RAM per node: 65855088Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 141ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 387
- Average Time: 125ms
- Event Batch Processing distribution:
	- 500ms: 356
	- 1000ms: 384
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
- Average Time: 176ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 451
- Average Time: 122ms
- Event Batch Processing distribution:
	- 500ms: 408
	- 1000ms: 450
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
- Average Time: 359ms
- Reload distribution:
	- 500ms: 779
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 406ms
- Event Batch Processing distribution:
	- 500ms: 683
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

- Total: 81
- Total Errors: 0
- Average Time: 126ms
- Reload distribution:
	- 500ms: 81
	- 1000ms: 81
	- 5000ms: 81
	- 10000ms: 81
	- 30000ms: 81
	- +Infms: 81

### Event Batch Processing

- Total: 83
- Average Time: 204ms
- Event Batch Processing distribution:
	- 500ms: 81
	- 1000ms: 83
	- 5000ms: 83
	- 10000ms: 83
	- 30000ms: 83
	- +Infms: 83

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
Requests      [total, rate, throughput]         30000, 1000.03, 998.69
Duration      [total, attack, wait]             30s, 29.999s, 964.397µs
Latencies     [min, mean, 50, 90, 95, 99, max]  509.471µs, 995.309µs, 961.016µs, 1.108ms, 1.171ms, 1.354ms, 12.787ms
Bytes In      [total, mean]                     4799883, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.87%
Status Codes  [code:count]                      200:29961  502:39  
Error Set:
502 Bad Gateway
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 958.326µs
Latencies     [min, mean, 50, 90, 95, 99, max]  719.622µs, 1.044ms, 1.023ms, 1.162ms, 1.21ms, 1.361ms, 22.043ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
