# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: e7d217a8f01fb3c8fc4507ef6f0e7feead667f20
- Date: 2024-11-14T18:42:55Z
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

- Total: 126
- Total Errors: 0
- Average Time: 288ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 385
- Average Time: 173ms
- Event Batch Processing distribution:
	- 500ms: 323
	- 1000ms: 382
	- 5000ms: 385
	- 10000ms: 385
	- 30000ms: 385
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
- Average Time: 367ms
- Reload distribution:
	- 500ms: 102
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 450
- Average Time: 172ms
- Event Batch Processing distribution:
	- 500ms: 378
	- 1000ms: 432
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
- Average Time: 2508ms
- Reload distribution:
	- 500ms: 79
	- 1000ms: 179
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 2578ms
- Event Batch Processing distribution:
	- 500ms: 80
	- 1000ms: 178
	- 5000ms: 996
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

- Total: 142
- Total Errors: 0
- Average Time: 151ms
- Reload distribution:
	- 500ms: 142
	- 1000ms: 142
	- 5000ms: 142
	- 10000ms: 142
	- 30000ms: 142
	- +Infms: 142

### Event Batch Processing

- Total: 145
- Average Time: 150ms
- Event Batch Processing distribution:
	- 500ms: 145
	- 1000ms: 145
	- 5000ms: 145
	- 10000ms: 145
	- 30000ms: 145
	- +Infms: 145

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
Requests      [total, rate, throughput]         30000, 1000.00, 995.95
Duration      [total, attack, wait]             30.001s, 30s, 652.553µs
Latencies     [min, mean, 50, 90, 95, 99, max]  307.059µs, 657.182µs, 636.168µs, 714.719µs, 747.948µs, 857.399µs, 18.835ms
Bytes In      [total, mean]                     4863388, 162.11
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.60%
Status Codes  [code:count]                      200:29879  503:121  
Error Set:
503 Service Temporarily Unavailable
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30s, 30s, 760.399µs
Latencies     [min, mean, 50, 90, 95, 99, max]  559.645µs, 714.017µs, 697.276µs, 795.75µs, 845.685µs, 946.55µs, 9.636ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
