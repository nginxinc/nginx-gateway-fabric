# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853984Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 4s
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

- Event Batch Total: 6
- Event Batch Processing Average Time: 71ms
- Event Batch Processing distribution:
	- 500.0ms: 6
	- 1000.0ms: 6
	- 5000.0ms: 6
	- 10000.0ms: 6
	- 30000.0ms: 6
	- +Infms: 6

## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 4s
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
- Event Batch Processing Average Time: 56ms
- Event Batch Processing distribution:
	- 500.0ms: 6
	- 1000.0ms: 6
	- 5000.0ms: 6
	- 10000.0ms: 6
	- 30000.0ms: 6
	- +Infms: 6

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 8s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 46
- NGINX Reload Average Time: 153ms
- Reload distribution:
	- 500.0ms: 46
	- 1000.0ms: 46
	- 5000.0ms: 46
	- 10000.0ms: 46
	- 30000.0ms: 46
	- +Infms: 46

### Event Batch Processing

- Event Batch Total: 321
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500.0ms: 321
	- 1000.0ms: 321
	- 5000.0ms: 321
	- 10000.0ms: 321
	- 30000.0ms: 321
	- +Infms: 321

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 44s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 245
- NGINX Reload Average Time: 153ms
- Reload distribution:
	- 500.0ms: 245
	- 1000.0ms: 245
	- 5000.0ms: 245
	- 10000.0ms: 245
	- 30000.0ms: 245
	- +Infms: 245

### Event Batch Processing

- Event Batch Total: 1597
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500.0ms: 1597
	- 1000.0ms: 1597
	- 5000.0ms: 1597
	- 10000.0ms: 1597
	- 30000.0ms: 1597
	- +Infms: 1597

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 48
- NGINX Reload Average Time: 149ms
- Reload distribution:
	- 500.0ms: 48
	- 1000.0ms: 48
	- 5000.0ms: 48
	- 10000.0ms: 48
	- 30000.0ms: 48
	- +Infms: 48

### Event Batch Processing

- Event Batch Total: 289
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500.0ms: 289
	- 1000.0ms: 289
	- 5000.0ms: 289
	- 10000.0ms: 289
	- 30000.0ms: 289
	- +Infms: 289

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 248
- NGINX Reload Average Time: 151ms
- Reload distribution:
	- 500.0ms: 248
	- 1000.0ms: 248
	- 5000.0ms: 248
	- 10000.0ms: 248
	- 30000.0ms: 248
	- +Infms: 248

### Event Batch Processing

- Event Batch Total: 1438
- Event Batch Processing Average Time: 30ms
- Event Batch Processing distribution:
	- 500.0ms: 1438
	- 1000.0ms: 1438
	- 5000.0ms: 1438
	- 10000.0ms: 1438
	- 30000.0ms: 1438
	- +Infms: 1438
