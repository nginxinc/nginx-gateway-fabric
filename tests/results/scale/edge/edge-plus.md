# Results

## Test environment

NGINX Plus: true

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

	- Total: 386
	- Average Time: 125ms
	- Event Batch Processing distribution:
		- 500ms: 351
		- 1000ms: 385
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
	- Average Time: 167ms
	- Reload distribution:
		- 500ms: 128
		- 1000ms: 128
		- 5000ms: 128
		- 10000ms: 128
		- 30000ms: 128
		- +Infms: 128

	### Event Batch Processing

	- Total: 450
	- Average Time: 117ms
	- Event Batch Processing distribution:
		- 500ms: 407
		- 1000ms: 449
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
	- Average Time: 348ms
	- Reload distribution:
		- 500ms: 834
		- 1000ms: 1001
		- 5000ms: 1001
		- 10000ms: 1001
		- 30000ms: 1001
		- +Infms: 1001

	### Event Batch Processing

	- Total: 1008
	- Average Time: 401ms
	- Event Batch Processing distribution:
		- 500ms: 708
		- 1000ms: 1007
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
	- Average Time: 126ms
	- Reload distribution:
		- 500ms: 3
		- 1000ms: 3
		- 5000ms: 3
		- 10000ms: 3
		- 30000ms: 3
		- +Infms: 3

	### Event Batch Processing

	- Total: 241
	- Average Time: 10ms
	- Event Batch Processing distribution:
		- 500ms: 241
		- 1000ms: 241
		- 5000ms: 241
		- 10000ms: 241
		- 30000ms: 241
		- +Infms: 241

	### Errors

	- NGF errors: 0
	- NGF container restarts: 0
	- NGINX errors: 2
	- NGINX container restarts: 0

	### Graphs and Logs

	See [output directory](./TestScale_UpstreamServers) for more details.
	The logs are attached only if there are errors.
	
## Test TestScale_HTTPMatches

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 773.622µs
Latencies     [min, mean, 50, 90, 95, 99, max]  573.137µs, 790.725µs, 760.739µs, 871.7µs, 916.717µs, 1.07ms, 12.522ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30s, 30s, 863.972µs
Latencies     [min, mean, 50, 90, 95, 99, max]  660.26µs, 868.747µs, 850.012µs, 992.647µs, 1.051ms, 1.164ms, 7.505ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
