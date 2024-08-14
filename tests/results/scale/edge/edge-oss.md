# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 9a85dbcc0797e31557a3731688795aa166ee0f96
- Date: 2024-08-13T21:12:05Z
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

- Total: 126
- Total Errors: 0
- Average Time: 150ms
- Reload distribution:
	- 500ms: 126
	- 1000ms: 126
	- 5000ms: 126
	- 10000ms: 126
	- 30000ms: 126
	- +Infms: 126

### Event Batch Processing

- Total: 385
- Average Time: 126ms
- Event Batch Processing distribution:
	- 500ms: 347
	- 1000ms: 385
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

- Total: 449
- Average Time: 121ms
- Event Batch Processing distribution:
	- 500ms: 407
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
- Average Time: 390ms
- Reload distribution:
	- 500ms: 699
	- 1000ms: 1001
	- 5000ms: 1001
	- 10000ms: 1001
	- 30000ms: 1001
	- +Infms: 1001

### Event Batch Processing

- Total: 1008
- Average Time: 442ms
- Event Batch Processing distribution:
	- 500ms: 611
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

- Total: 150
- Total Errors: 0
- Average Time: 127ms
- Reload distribution:
	- 500ms: 150
	- 1000ms: 150
	- 5000ms: 150
	- 10000ms: 150
	- 30000ms: 150
	- +Infms: 150

### Event Batch Processing

- Total: 153
- Average Time: 126ms
- Event Batch Processing distribution:
	- 500ms: 153
	- 1000ms: 153
	- 5000ms: 153
	- 10000ms: 153
	- 30000ms: 153
	- +Infms: 153

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
Requests      [total, rate, throughput]         30000, 1000.02, 997.90
Duration      [total, attack, wait]             30s, 29.999s, 725.405µs
Latencies     [min, mean, 50, 90, 95, 99, max]  350.433µs, 744.47µs, 721.353µs, 821.981µs, 863.431µs, 986.153µs, 12.411ms
Bytes In      [total, mean]                     4799370, 159.98
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.79%
Status Codes  [code:count]                      200:29937  502:63  
Error Set:
502 Bad Gateway
```
```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 727.319µs
Latencies     [min, mean, 50, 90, 95, 99, max]  616.109µs, 832.616µs, 810.194µs, 948.088µs, 1.001ms, 1.127ms, 12.141ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
