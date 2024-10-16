# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 3a08fdafadfe0fb4a9c25679da1a1fcd6b181474
- Date: 2024-10-15T13:45:52Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1014001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 4s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 88ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 71ms
- Event Batch Processing distribution:
	- 500ms: 6
	- 1000ms: 6
	- 5000ms: 6
	- 10000ms: 6
	- 30000ms: 6
	- +Infms: 6

## Test 1: Resources exist before startup - NumResources 150

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
- Event Batch Processing Average Time: 47ms
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
- NGINX Reloads: 344
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 344
	- 1000ms: 344
	- 5000ms: 344
	- 10000ms: 344
	- 30000ms: 344
	- +Infms: 344

### Event Batch Processing

- Event Batch Total: 1542
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1542
	- 1000ms: 1542
	- 5000ms: 1542
	- 10000ms: 1542
	- 30000ms: 1542
	- +Infms: 1542
