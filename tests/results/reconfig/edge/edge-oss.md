# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855004Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 3s
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
- Event Batch Processing Average Time: 48ms
- Event Batch Processing distribution:
	- 500ms: 6
	- 1000ms: 6
	- 5000ms: 6
	- 10000ms: 6
	- 30000ms: 6
	- +Infms: 6

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 7s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 62
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 62
	- 1000ms: 62
	- 5000ms: 62
	- 10000ms: 62
	- 30000ms: 62
	- +Infms: 62

### Event Batch Processing

- Event Batch Total: 336
- Event Batch Processing Average Time: 23ms
- Event Batch Processing distribution:
	- 500ms: 336
	- 1000ms: 336
	- 5000ms: 336
	- 10000ms: 336
	- 30000ms: 336
	- +Infms: 336

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 343
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 343
	- 1000ms: 343
	- 5000ms: 343
	- 10000ms: 343
	- 30000ms: 343
	- +Infms: 343

### Event Batch Processing

- Event Batch Total: 1696
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1696
	- 1000ms: 1696
	- 5000ms: 1696
	- 10000ms: 1696
	- 30000ms: 1696
	- +Infms: 1696

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 64
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 64
	- 1000ms: 64
	- 5000ms: 64
	- 10000ms: 64
	- 30000ms: 64
	- +Infms: 64

### Event Batch Processing

- Event Batch Total: 304
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 304
	- 1000ms: 304
	- 5000ms: 304
	- 10000ms: 304
	- 30000ms: 304
	- +Infms: 304

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
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

- Event Batch Total: 1547
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1547
	- 1000ms: 1547
	- 5000ms: 1547
	- 10000ms: 1547
	- 30000ms: 1547
	- +Infms: 1547
