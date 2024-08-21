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

- Total: 386
- Average Time: 134ms
- Event Batch Processing distribution:
	- 500ms: 350
	- 1000ms: 386
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

- Total: 126
- Total Errors: 0
- Average Time: 201ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 448
- Average Time: 129ms
- Event Batch Processing distribution:
	- 500ms: 401
	- 1000ms: 446
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
- Average Time: 894ms
- Reload distribution:
	- 500ms: 442
	- 1000ms: 592
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1007
- Average Time: 1052ms
- Event Batch Processing distribution:
	- 500ms: 378
	- 1000ms: 597
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
- Average Time: 127ms
- Reload distribution:
	- 500ms: 3
	- 1000ms: 3
	- 5000ms: 3
	- 10000ms: 3
	- 30000ms: 3
	- +Infms: 3

### Event Batch Processing

- Total: 174
- Average Time: 52ms
- Event Batch Processing distribution:
	- 500ms: 174
	- 1000ms: 174
	- 5000ms: 174
	- 10000ms: 174
	- 30000ms: 174
	- +Infms: 174

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
Requests      [total, rate, throughput]         30000, 1000.00, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 873.367µs
Latencies     [min, mean, 50, 90, 95, 99, max]  631.55µs, 1.294ms, 1.005ms, 1.548ms, 2.248ms, 8.757ms, 23.886ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.96
Duration      [total, attack, wait]             30.001s, 30s, 1.218ms
Latencies     [min, mean, 50, 90, 95, 99, max]  735.98µs, 4.022ms, 1.2ms, 2.683ms, 8.838ms, 76.295ms, 305.702ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
