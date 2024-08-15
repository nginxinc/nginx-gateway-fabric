# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3c029b1417c1f89f2a29aeef07f47078640e28b2
- Date: 2024-08-15T00:04:25Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1326000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test TestScale_Listeners

### Reloads

- Total: 128
- Total Errors: 0
- Average Time: 151ms
- Reload distribution:
	- 500ms: 128
	- 1000ms: 128
	- 5000ms: 128
	- 10000ms: 128
	- 30000ms: 128
	- +Infms: 128

### Event Batch Processing

- Total: 386
- Average Time: 124ms
- Event Batch Processing distribution:
	- 500ms: 352
	- 1000ms: 384
	- 5000ms: 386
	- 10000ms: 386
	- 30000ms: 386
	- +Infms: 386

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
- Average Time: 177ms
- Reload distribution:
	- 500ms: 127
	- 1000ms: 127
	- 5000ms: 127
	- 10000ms: 127
	- 30000ms: 127
	- +Infms: 127

### Event Batch Processing

- Total: 448
- Average Time: 120ms
- Event Batch Processing distribution:
	- 500ms: 402
	- 1000ms: 447
	- 5000ms: 448
	- 10000ms: 448
	- 30000ms: 448
	- +Infms: 448

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
- Average Time: 394ms
- Reload distribution:
	- 500ms: 687
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 444ms
- Event Batch Processing distribution:
	- 500ms: 613
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

- Total: 107
- Total Errors: 0
- Average Time: 126ms
- Reload distribution:
	- 500ms: 107
	- 1000ms: 107
	- 5000ms: 107
	- 10000ms: 107
	- 30000ms: 107
	- +Infms: 107

### Event Batch Processing

- Total: 110
- Average Time: 125ms
- Event Batch Processing distribution:
	- 500ms: 110
	- 1000ms: 110
	- 5000ms: 110
	- 10000ms: 110
	- 30000ms: 110
	- +Infms: 110

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
Duration      [total, attack, wait]             30.001s, 30s, 663.08µs
Latencies     [min, mean, 50, 90, 95, 99, max]  518.253µs, 721.379µs, 690.364µs, 801.458µs, 846.06µs, 972.256µs, 12.432ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 775.127µs
Latencies     [min, mean, 50, 90, 95, 99, max]  586.924µs, 794.034µs, 771.093µs, 911.524µs, 965.023µs, 1.086ms, 12.692ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
