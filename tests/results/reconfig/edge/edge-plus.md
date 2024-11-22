# Results

## Test environment

NGINX Plus: true

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
- NGINX Reload Average Time: 114ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 5
- Event Batch Processing Average Time: 56ms
- Event Batch Processing distribution:
	- 500ms: 5
	- 1000ms: 5
	- 5000ms: 5
	- 10000ms: 5
	- 30000ms: 5
	- +Infms: 5

## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
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

- Event Batch Total: 337
- Event Batch Processing Average Time: 23ms
- Event Batch Processing distribution:
	- 500ms: 337
	- 1000ms: 337
	- 5000ms: 337
	- 10000ms: 337
	- 30000ms: 337
	- +Infms: 337

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 338
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 338
	- 1000ms: 338
	- 5000ms: 338
	- 10000ms: 338
	- 30000ms: 338
	- +Infms: 338

### Event Batch Processing

- Event Batch Total: 1693
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1693
	- 1000ms: 1693
	- 5000ms: 1693
	- 10000ms: 1693
	- 30000ms: 1693
	- +Infms: 1693

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

- Event Batch Total: 306
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 306
	- 1000ms: 306
	- 5000ms: 306
	- 10000ms: 306
	- 30000ms: 306
	- +Infms: 306

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 342
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 342
	- 1000ms: 342
	- 5000ms: 342
	- 10000ms: 342
	- 30000ms: 342
	- +Infms: 342

### Event Batch Processing

- Event Batch Total: 1534
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1534
	- 1000ms: 1534
	- 5000ms: 1534
	- 10000ms: 1534
	- 30000ms: 1534
	- +Infms: 1534
