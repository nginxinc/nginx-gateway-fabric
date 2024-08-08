# Results

## Test environment

NGINX Plus: false

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

	- Total: 126
	- Total Errors: 0
	- Average Time: 149ms
	- Reload distribution:
		- 500ms: 126
		- 1000ms: 126
		- 5000ms: 126
		- 10000ms: 126
		- 30000ms: 126
		- +Infms: 126

	### Event Batch Processing

	- Total: 384
	- Average Time: 126ms
	- Event Batch Processing distribution:
		- 500ms: 347
		- 1000ms: 381
		- 5000ms: 384
		- 10000ms: 384
		- 30000ms: 384
		- +Infms: 384

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
	- Average Time: 178ms
	- Reload distribution:
		- 500ms: 128
		- 1000ms: 128
		- 5000ms: 128
		- 10000ms: 128
		- 30000ms: 128
		- +Infms: 128

	### Event Batch Processing

	- Total: 450
	- Average Time: 120ms
	- Event Batch Processing distribution:
		- 500ms: 405
		- 1000ms: 450
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
	- Average Time: 396ms
	- Reload distribution:
		- 500ms: 688
		- 1000ms: 1001
		- 5000ms: 1001
		- 10000ms: 1001
		- 30000ms: 1001
		- +Infms: 1001

	### Event Batch Processing

	- Total: 1008
	- Average Time: 446ms
	- Event Batch Processing distribution:
		- 500ms: 599
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

	- Total: 158
	- Total Errors: 0
	- Average Time: 127ms
	- Reload distribution:
		- 500ms: 158
		- 1000ms: 158
		- 5000ms: 158
		- 10000ms: 158
		- 30000ms: 158
		- +Infms: 158

	### Event Batch Processing

	- Total: 161
	- Average Time: 126ms
	- Event Batch Processing distribution:
		- 500ms: 161
		- 1000ms: 161
		- 5000ms: 161
		- 10000ms: 161
		- 30000ms: 161
		- +Infms: 161

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
Requests      [total, rate, throughput]         30000, 1000.03, 999.07
Duration      [total, attack, wait]             30s, 29.999s, 710.713µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.568µs, 802.426µs, 763.265µs, 882.309µs, 929.834µs, 1.082ms, 15.729ms
Bytes In      [total, mean]                     4829692, 160.99
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.91%
Status Codes  [code:count]                      200:29972  502:28  
Error Set:
502 Bad Gateway
```
```text
Requests      [total, rate, throughput]         30000, 1000.06, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 900.68µs
Latencies     [min, mean, 50, 90, 95, 99, max]  626.791µs, 890.845µs, 863.82µs, 1.003ms, 1.063ms, 1.216ms, 17.624ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
