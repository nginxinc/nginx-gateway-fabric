# Results

## Test environment

NGINX Plus: true

 NGINX Gateway Fabric:

- Commit: unknown
- Date: unknown
- Dirty: unknown

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
	- Average Time: 146ms
	- Reload distribution:
		- 500ms: 127
		- 1000ms: 127
		- 5000ms: 127
		- 10000ms: 127
		- 30000ms: 127
		- +Infms: 127

	### Event Batch Processing

	- Total: 385
	- Average Time: 125ms
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

	- Total: 126
	- Total Errors: 0
	- Average Time: 170ms
	- Reload distribution:
		- 500ms: 126
		- 1000ms: 126
		- 5000ms: 126
		- 10000ms: 126
		- 30000ms: 126
		- +Infms: 126

	### Event Batch Processing

	- Total: 448
	- Average Time: 119ms
	- Event Batch Processing distribution:
		- 500ms: 404
		- 1000ms: 448
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
	- Average Time: 363ms
	- Reload distribution:
		- 500ms: 784
		- 1000ms: 1001
		- 5000ms: 1001
		- 10000ms: 1001
		- 30000ms: 1001
		- +Infms: 1001

	### Event Batch Processing

	- Total: 1008
	- Average Time: 421ms
	- Event Batch Processing distribution:
		- 500ms: 667
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
	- Average Time: 134ms
	- Reload distribution:
		- 500ms: 3
		- 1000ms: 3
		- 5000ms: 3
		- 10000ms: 3
		- 30000ms: 3
		- +Infms: 3

	### Event Batch Processing

	- Total: 262
	- Average Time: 11ms
	- Event Batch Processing distribution:
		- 500ms: 262
		- 1000ms: 262
		- 5000ms: 262
		- 10000ms: 262
		- 30000ms: 262
		- +Infms: 262

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
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 620.897µs
Latencies     [min, mean, 50, 90, 95, 99, max]  514.425µs, 661.253µs, 648.826µs, 733.192µs, 766.876µs, 851.306µs, 8.698ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.06, 1000.04
Duration      [total, attack, wait]             29.999s, 29.998s, 824.534µs
Latencies     [min, mean, 50, 90, 95, 99, max]  581.605µs, 737.731µs, 719.088µs, 823.825µs, 872.733µs, 968.365µs, 10.977ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
