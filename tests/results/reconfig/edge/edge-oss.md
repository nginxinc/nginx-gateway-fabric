# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853980Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 3s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 101ms
- Reload distribution:
	- 500.0ms: 2
	- 1000.0ms: 2
	- 5000.0ms: 2
	- 10000.0ms: 2
	- 30000.0ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 5
- Event Batch Processing Average Time: 58ms
- Event Batch Processing distribution:
	- 500.0ms: 5
	- 1000.0ms: 5
	- 5000.0ms: 5
	- 10000.0ms: 5
	- 30000.0ms: 5
	- +Infms: 5

## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500.0ms: 2
	- 1000.0ms: 2
	- 5000.0ms: 2
	- 10000.0ms: 2
	- 30000.0ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 51ms
- Event Batch Processing distribution:
	- 500.0ms: 6
	- 1000.0ms: 6
	- 5000.0ms: 6
	- 10000.0ms: 6
	- 30000.0ms: 6
	- +Infms: 6

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 7s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 53
- NGINX Reload Average Time: 149ms
- Reload distribution:
	- 500.0ms: 53
	- 1000.0ms: 53
	- 5000.0ms: 53
	- 10000.0ms: 53
	- 30000.0ms: 53
	- +Infms: 53

### Event Batch Processing

- Event Batch Total: 328
- Event Batch Processing Average Time: 24ms
- Event Batch Processing distribution:
	- 500.0ms: 328
	- 1000.0ms: 328
	- 5000.0ms: 328
	- 10000.0ms: 328
	- 30000.0ms: 328
	- +Infms: 328

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 288
- NGINX Reload Average Time: 150ms
- Reload distribution:
	- 500.0ms: 288
	- 1000.0ms: 288
	- 5000.0ms: 288
	- 10000.0ms: 288
	- 30000.0ms: 288
	- +Infms: 288

### Event Batch Processing

- Event Batch Total: 1641
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500.0ms: 1641
	- 1000.0ms: 1641
	- 5000.0ms: 1641
	- 10000.0ms: 1641
	- 30000.0ms: 1641
	- +Infms: 1641

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 61
- NGINX Reload Average Time: 132ms
- Reload distribution:
	- 500.0ms: 61
	- 1000.0ms: 61
	- 5000.0ms: 61
	- 10000.0ms: 61
	- 30000.0ms: 61
	- +Infms: 61

### Event Batch Processing

- Event Batch Total: 305
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500.0ms: 305
	- 1000.0ms: 305
	- 5000.0ms: 305
	- 10000.0ms: 305
	- 30000.0ms: 305
	- +Infms: 305

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 327
- NGINX Reload Average Time: 132ms
- Reload distribution:
	- 500.0ms: 327
	- 1000.0ms: 327
	- 5000.0ms: 327
	- 10000.0ms: 327
	- 30000.0ms: 327
	- +Infms: 327

### Event Batch Processing

- Event Batch Total: 1539
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500.0ms: 1539
	- 1000.0ms: 1539
	- 5000.0ms: 1539
	- 10000.0ms: 1539
	- 30000.0ms: 1539
	- +Infms: 1539
