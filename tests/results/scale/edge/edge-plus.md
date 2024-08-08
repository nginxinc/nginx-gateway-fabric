# Results

## Test environment

NGINX Plus: true

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

	- Total: 127
	- Total Errors: 0
	- Average Time: 148ms
	- Reload distribution:
		- 500ms: 127
		- 1000ms: 127
		- 5000ms: 127
		- 10000ms: 127
		- 30000ms: 127
		- +Infms: 127

	### Event Batch Processing

	- Total: 386
	- Average Time: 128ms
	- Event Batch Processing distribution:
		- 500ms: 351
		- 1000ms: 383
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

	- Total: 128
	- Total Errors: 0
	- Average Time: 171ms
	- Reload distribution:
		- 500ms: 128
		- 1000ms: 128
		- 5000ms: 128
		- 10000ms: 128
		- 30000ms: 128
		- +Infms: 128

	### Event Batch Processing

	- Total: 449
	- Average Time: 124ms
	- Event Batch Processing distribution:
		- 500ms: 407
		- 1000ms: 445
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
	- Average Time: 354ms
	- Reload distribution:
		- 500ms: 806
		- 1000ms: 1001
		- 5000ms: 1001
		- 10000ms: 1001
		- 30000ms: 1001
		- +Infms: 1001

	### Event Batch Processing

	- Total: 1007
	- Average Time: 409ms
	- Event Batch Processing distribution:
		- 500ms: 697
		- 1000ms: 1007
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

	- Total: 177
	- Average Time: 11ms
	- Event Batch Processing distribution:
		- 500ms: 177
		- 1000ms: 177
		- 5000ms: 177
		- 10000ms: 177
		- 30000ms: 177
		- +Infms: 177

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
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 701.792µs
Latencies     [min, mean, 50, 90, 95, 99, max]  518.88µs, 746.108µs, 709.034µs, 829.262µs, 876.982µs, 1.023ms, 23.083ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 677.445µs
Latencies     [min, mean, 50, 90, 95, 99, max]  597.855µs, 817.14µs, 797.937µs, 938.117µs, 994.676µs, 1.117ms, 11.675ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
