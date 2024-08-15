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

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 2s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 113ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 48ms
- Event Batch Processing distribution:
	- 500ms: 6
	- 1000ms: 6
	- 5000ms: 6
	- 10000ms: 6
	- 30000ms: 6
	- +Infms: 6


## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 2s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 113ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 46ms
- Event Batch Processing distribution:
	- 500ms: 6
	- 1000ms: 6
	- 5000ms: 6
	- 10000ms: 6
	- 30000ms: 6
	- +Infms: 6


## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 8s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 62
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 62
	- 1000ms: 62
	- 5000ms: 62
	- 10000ms: 62
	- 30000ms: 62
	- +Infms: 62

### Event Batch Processing

- Event Batch Total: 338
- Event Batch Processing Average Time: 23ms
- Event Batch Processing distribution:
	- 500ms: 338
	- 1000ms: 338
	- 5000ms: 338
	- 10000ms: 338
	- 30000ms: 338
	- +Infms: 338


## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 44s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 341
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 341
	- 1000ms: 341
	- 5000ms: 341
	- 10000ms: 341
	- 30000ms: 341
	- +Infms: 341

### Event Batch Processing

- Event Batch Total: 1695
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1695
	- 1000ms: 1695
	- 5000ms: 1695
	- 10000ms: 1695
	- 30000ms: 1695
	- +Infms: 1695


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 63
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 63
	- 1000ms: 63
	- 5000ms: 63
	- 10000ms: 63
	- 30000ms: 63
	- +Infms: 63

### Event Batch Processing

- Event Batch Total: 307
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 307
	- 1000ms: 307
	- 5000ms: 307
	- 10000ms: 307
	- 30000ms: 307
	- +Infms: 307


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 345
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 345
	- 1000ms: 345
	- 5000ms: 345
	- 10000ms: 345
	- 30000ms: 345
	- +Infms: 345

### Event Batch Processing

- Event Batch Total: 1547
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1547
	- 1000ms: 1547
	- 5000ms: 1547
	- 10000ms: 1547
	- 30000ms: 1547
	- +Infms: 1547

