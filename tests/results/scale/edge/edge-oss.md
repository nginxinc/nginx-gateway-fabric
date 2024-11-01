# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: fed4239ecb35f937b66bba7bd68d6894ca0762b3
- Date: 2024-11-01T00:13:12Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1355000
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 127
- Total Errors: 0
- Average Time: 283ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 385
- Average Time: 169ms
- Event Batch Processing distribution:
	- 500ms: 331
	- 1000ms: 380
	- 5000ms: 385
	- 10000ms: 385
	- 30000ms: 385
	- +Infms: 385

### Errors

- NGF errors: 1
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
- Average Time: 375ms
- Reload distribution:
	- 500ms: 97
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 451
- Average Time: 178ms
- Event Batch Processing distribution:
	- 500ms: 375
	- 1000ms: 428
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
- Average Time: 2648ms
- Reload distribution:
	- 500ms: 76
	- 1000ms: 176
	- 5000ms: 963
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 2722ms
- Event Batch Processing distribution:
	- 500ms: 77
	- 1000ms: 172
	- 5000ms: 948
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

- Total: 120
- Total Errors: 0
- Average Time: 151ms
- Reload distribution:
	- 500ms: 120
	- 1000ms: 120
	- 5000ms: 120
	- 10000ms: 120
	- 30000ms: 120
	- +Infms: 120

### Event Batch Processing

- Total: 123
- Average Time: 149ms
- Event Batch Processing distribution:
	- 500ms: 123
	- 1000ms: 123
	- 5000ms: 123
	- 10000ms: 123
	- 30000ms: 123
	- +Infms: 123

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
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 629.241µs
Latencies     [min, mean, 50, 90, 95, 99, max]  543.665µs, 721.615µs, 706.252µs, 795.418µs, 829.703µs, 936.424µs, 12.183ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 816.8µs
Latencies     [min, mean, 50, 90, 95, 99, max]  607.379µs, 793.307µs, 779.366µs, 885.14µs, 928.75µs, 1.035ms, 10.825ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
