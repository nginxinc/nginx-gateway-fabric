# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 809c0838e2f2658c3c4cd48325ffb0bc5a92a002
- Date: 2024-08-08T18:03:35Z
- Dirty: false

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
	- Average Time: 123ms
	- Event Batch Processing distribution:
		- 500ms: 353
		- 1000ms: 387
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
	- Average Time: 117ms
	- Event Batch Processing distribution:
		- 500ms: 405
		- 1000ms: 451
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
	- Average Time: 392ms
	- Reload distribution:
		- 500ms: 693
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

	- Total: 140
	- Total Errors: 0
	- Average Time: 129ms
	- Reload distribution:
		- 500ms: 140
		- 1000ms: 140
		- 5000ms: 140
		- 10000ms: 140
		- 30000ms: 140
		- +Infms: 140

	### Event Batch Processing

	- Total: 143
	- Average Time: 128ms
	- Event Batch Processing distribution:
		- 500ms: 143
		- 1000ms: 143
		- 5000ms: 143
		- 10000ms: 143
		- 30000ms: 143
		- +Infms: 143

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
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 855.501µs
Latencies     [min, mean, 50, 90, 95, 99, max]  541.471µs, 788.395µs, 754.567µs, 884.212µs, 933.098µs, 1.091ms, 18.578ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 779.338µs
Latencies     [min, mean, 50, 90, 95, 99, max]  626.365µs, 862.529µs, 838.648µs, 991.736µs, 1.054ms, 1.192ms, 12.484ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
