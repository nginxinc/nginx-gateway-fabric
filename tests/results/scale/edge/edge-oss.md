# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1254000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

	## Test TestScale_Listeners

	### Reloads

	- Total: 128
	- Total Errors: 0
	- Average Time: 147ms
	- Reload distribution:
		- 500ms: 128
		- 1000ms: 128
		- 5000ms: 128
		- 10000ms: 128
		- 30000ms: 128
		- +Infms: 128

	### Event Batch Processing

	- Total: 387
	- Average Time: 124ms
	- Event Batch Processing distribution:
		- 500ms: 352
		- 1000ms: 386
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
	- Average Time: 124ms
	- Event Batch Processing distribution:
		- 500ms: 405
		- 1000ms: 448
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
	- Average Time: 396ms
	- Reload distribution:
		- 500ms: 683
		- 1000ms: 1001
		- 5000ms: 1001
		- 10000ms: 1001
		- 30000ms: 1001
		- +Infms: 1001

	### Event Batch Processing

	- Total: 1008
	- Average Time: 451ms
	- Event Batch Processing distribution:
		- 500ms: 592
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

	- Total: 188
	- Total Errors: 0
	- Average Time: 127ms
	- Reload distribution:
		- 500ms: 188
		- 1000ms: 188
		- 5000ms: 188
		- 10000ms: 188
		- 30000ms: 188
		- +Infms: 188

	### Event Batch Processing

	- Total: 191
	- Average Time: 127ms
	- Event Batch Processing distribution:
		- 500ms: 191
		- 1000ms: 191
		- 5000ms: 191
		- 10000ms: 191
		- 30000ms: 191
		- +Infms: 191

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
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 665.948µs
Latencies     [min, mean, 50, 90, 95, 99, max]  537.268µs, 725.23µs, 700.355µs, 785.759µs, 819.079µs, 924.248µs, 19.946ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 789.991µs
Latencies     [min, mean, 50, 90, 95, 99, max]  617.838µs, 792.952µs, 777.13µs, 893.249µs, 933.87µs, 1.026ms, 9.302ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
