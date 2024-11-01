# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: fed4239ecb35f937b66bba7bd68d6894ca0762b3
- Date: 2024-11-01T00:13:12Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1355000
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 129ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 56ms
- Event Batch Processing distribution:
	- 500ms: 6
	- 1000ms: 6
	- 5000ms: 6
	- 10000ms: 6
	- 30000ms: 6
	- +Infms: 6

## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 51ms
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
- NGINX Reloads: 52
- NGINX Reload Average Time: 150ms
- Reload distribution:
	- 500ms: 52
	- 1000ms: 52
	- 5000ms: 52
	- 10000ms: 52
	- 30000ms: 52
	- +Infms: 52

### Event Batch Processing

- Event Batch Total: 327
- Event Batch Processing Average Time: 24ms
- Event Batch Processing distribution:
	- 500ms: 327
	- 1000ms: 327
	- 5000ms: 327
	- 10000ms: 327
	- 30000ms: 327
	- +Infms: 327

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 284
- NGINX Reload Average Time: 152ms
- Reload distribution:
	- 500ms: 284
	- 1000ms: 284
	- 5000ms: 284
	- 10000ms: 284
	- 30000ms: 284
	- +Infms: 284

### Event Batch Processing

- Event Batch Total: 1637
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 1637
	- 1000ms: 1637
	- 5000ms: 1637
	- 10000ms: 1637
	- 30000ms: 1637
	- +Infms: 1637

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 54
- NGINX Reload Average Time: 149ms
- Reload distribution:
	- 500ms: 54
	- 1000ms: 54
	- 5000ms: 54
	- 10000ms: 54
	- 30000ms: 54
	- +Infms: 54

### Event Batch Processing

- Event Batch Total: 295
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500ms: 295
	- 1000ms: 295
	- 5000ms: 295
	- 10000ms: 295
	- 30000ms: 295
	- +Infms: 295

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 296
- NGINX Reload Average Time: 147ms
- Reload distribution:
	- 500ms: 296
	- 1000ms: 296
	- 5000ms: 296
	- 10000ms: 296
	- 30000ms: 296
	- +Infms: 296

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
