# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3a08fdafadfe0fb4a9c25679da1a1fcd6b181474
- Date: 2024-10-15T13:45:52Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1014001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 287ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 385
- Average Time: 168ms
- Event Batch Processing distribution:
	- 500ms: 326
	- 1000ms: 380
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
- Average Time: 369ms
- Reload distribution:
	- 500ms: 100
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 450
- Average Time: 172ms
- Event Batch Processing distribution:
	- 500ms: 375
	- 1000ms: 434
	- 5000ms: 450
	- 10000ms: 450
	- 30000ms: 450
	- +Infms: 450

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
- Average Time: 2645ms
- Reload distribution:
	- 500ms: 77
	- 1000ms: 178
	- 5000ms: 961
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 2715ms
- Event Batch Processing distribution:
	- 500ms: 76
	- 1000ms: 176
	- 5000ms: 950
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

- Total: 168
- Total Errors: 0
- Average Time: 151ms
- Reload distribution:
	- 500ms: 168
	- 1000ms: 168
	- 5000ms: 168
	- 10000ms: 168
	- 30000ms: 168
	- +Infms: 168

### Event Batch Processing

- Total: 170
- Average Time: 151ms
- Event Batch Processing distribution:
	- 500ms: 170
	- 1000ms: 170
	- 5000ms: 170
	- 10000ms: 170
	- 30000ms: 170
	- +Infms: 170

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
Requests      [total, rate, throughput]         30000, 1000.03, 997.17
Duration      [total, attack, wait]             30s, 29.999s, 729.152µs
Latencies     [min, mean, 50, 90, 95, 99, max]  350.617µs, 745.204µs, 729.326µs, 835.368µs, 875.015µs, 982.22µs, 13.169ms
Bytes In      [total, mean]                     4829065, 160.97
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.72%
Status Codes  [code:count]                      200:29915  502:85  
Error Set:
502 Bad Gateway
```
```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 787.081µs
Latencies     [min, mean, 50, 90, 95, 99, max]  599.493µs, 831.941µs, 815.271µs, 941.854µs, 991.59µs, 1.115ms, 7.879ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
