# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 149ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 386
- Average Time: 124ms
- Event Batch Processing distribution:
	- 500ms: 352
	- 1000ms: 386
	- 5000ms: 386
	- 10000ms: 386
	- 30000ms: 386
	- +Infms: 386

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

- Total: 127
- Total Errors: 0
- Average Time: 170ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 449
- Average Time: 120ms
- Event Batch Processing distribution:
	- 500ms: 407
	- 1000ms: 447
	- 5000ms: 449
	- 10000ms: 449
	- 30000ms: 449
	- +Infms: 449

### Errors

- NGF errors: 3
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
- Average Time: 361ms
- Reload distribution:
	- 500ms: 790
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 424ms
- Event Batch Processing distribution:
	- 500ms: 656
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

- Total: 143
- Average Time: 8ms
- Event Batch Processing distribution:
	- 500ms: 143
	- 1000ms: 143
	- 5000ms: 143
	- 10000ms: 143
	- 30000ms: 143
	- +Infms: 143

### Errors

- NGF errors: 1
- NGF container restarts: 0
- NGINX errors: 0
- NGINX container restarts: 0

### Graphs and Logs

See [output directory](./TestScale_UpstreamServers) for more details.
The logs are attached only if there are errors.

## Test TestScale_HTTPMatches

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 859.047µs
Latencies     [min, mean, 50, 90, 95, 99, max]  514.844µs, 721.442µs, 685.71µs, 795.965µs, 841.032µs, 978.573µs, 22.353ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 882.046µs
Latencies     [min, mean, 50, 90, 95, 99, max]  600.374µs, 792.586µs, 764.498µs, 897.846µs, 945.622µs, 1.063ms, 16.397ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
