# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: e7d217a8f01fb3c8fc4507ef6f0e7feead667f20
- Date: 2024-11-14T18:42:55Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 3s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 5
- Event Batch Processing Average Time: 60ms
- Event Batch Processing distribution:
	- 500ms: 5
	- 1000ms: 5
	- 5000ms: 5
	- 10000ms: 5
	- 30000ms: 5
	- +Infms: 5

## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 2s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 50ms
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
- NGINX Reloads: 52
- NGINX Reload Average Time: 151ms
- Reload distribution:
	- 500ms: 52
	- 1000ms: 52
	- 5000ms: 52
	- 10000ms: 52
	- 30000ms: 52
	- +Infms: 52

### Event Batch Processing

- Event Batch Total: 326
- Event Batch Processing Average Time: 24ms
- Event Batch Processing distribution:
	- 500ms: 326
	- 1000ms: 326
	- 5000ms: 326
	- 10000ms: 326
	- 30000ms: 326
	- +Infms: 326

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 283
- NGINX Reload Average Time: 152ms
- Reload distribution:
	- 500ms: 283
	- 1000ms: 283
	- 5000ms: 283
	- 10000ms: 283
	- 30000ms: 283
	- +Infms: 283

### Event Batch Processing

- Event Batch Total: 1633
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 1633
	- 1000ms: 1633
	- 5000ms: 1633
	- 10000ms: 1633
	- 30000ms: 1633
	- +Infms: 1633

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 55
- NGINX Reload Average Time: 148ms
- Reload distribution:
	- 500ms: 55
	- 1000ms: 55
	- 5000ms: 55
	- 10000ms: 55
	- 30000ms: 55
	- +Infms: 55

### Event Batch Processing

- Event Batch Total: 296
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500ms: 296
	- 1000ms: 296
	- 5000ms: 296
	- 10000ms: 296
	- 30000ms: 296
	- +Infms: 296

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 294
- NGINX Reload Average Time: 148ms
- Reload distribution:
	- 500ms: 294
	- 1000ms: 294
	- 5000ms: 294
	- 10000ms: 294
	- 30000ms: 294
	- +Infms: 294

### Event Batch Processing

- Event Batch Total: 1506
- Event Batch Processing Average Time: 29ms
- Event Batch Processing distribution:
	- 500ms: 1506
	- 1000ms: 1506
	- 5000ms: 1506
	- 10000ms: 1506
	- 30000ms: 1506
	- +Infms: 1506
