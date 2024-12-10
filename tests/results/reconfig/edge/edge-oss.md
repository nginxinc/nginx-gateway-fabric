# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 929413c15af7bee3adb32e103c9d1513a693da16
- Date: 2024-11-28T12:52:45Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853964Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

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
- Event Batch Processing Average Time: 50ms
- Event Batch Processing distribution:
	- 500.0ms: 6
	- 1000.0ms: 6
	- 5000.0ms: 6
	- 10000.0ms: 6
	- 30000.0ms: 6
	- +Infms: 6

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
- Event Batch Processing Average Time: 59ms
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
- NGINX Reloads: 52
- NGINX Reload Average Time: 149ms
- Reload distribution:
	- 500.0ms: 52
	- 1000.0ms: 52
	- 5000.0ms: 52
	- 10000.0ms: 52
	- 30000.0ms: 52
	- +Infms: 52

### Event Batch Processing

- Event Batch Total: 327
- Event Batch Processing Average Time: 24ms
- Event Batch Processing distribution:
	- 500.0ms: 327
	- 1000.0ms: 327
	- 5000.0ms: 327
	- 10000.0ms: 327
	- 30000.0ms: 327
	- +Infms: 327

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 284
- NGINX Reload Average Time: 152ms
- Reload distribution:
	- 500.0ms: 284
	- 1000.0ms: 284
	- 5000.0ms: 284
	- 10000.0ms: 284
	- 30000.0ms: 284
	- +Infms: 284

### Event Batch Processing

- Event Batch Total: 1637
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500.0ms: 1637
	- 1000.0ms: 1637
	- 5000.0ms: 1637
	- 10000.0ms: 1637
	- 30000.0ms: 1637
	- +Infms: 1637

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 54
- NGINX Reload Average Time: 150ms
- Reload distribution:
	- 500.0ms: 54
	- 1000.0ms: 54
	- 5000.0ms: 54
	- 10000.0ms: 54
	- 30000.0ms: 54
	- +Infms: 54

### Event Batch Processing

- Event Batch Total: 296
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500.0ms: 296
	- 1000.0ms: 296
	- 5000.0ms: 296
	- 10000.0ms: 296
	- 30000.0ms: 296
	- +Infms: 296

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 290
- NGINX Reload Average Time: 150ms
- Reload distribution:
	- 500.0ms: 290
	- 1000.0ms: 290
	- 5000.0ms: 290
	- 10000.0ms: 290
	- 30000.0ms: 290
	- +Infms: 290

### Event Batch Processing

- Event Batch Total: 1487
- Event Batch Processing Average Time: 29ms
- Event Batch Processing distribution:
	- 500.0ms: 1487
	- 1000.0ms: 1487
	- 5000.0ms: 1487
	- 10000.0ms: 1487
	- 30000.0ms: 1487
	- +Infms: 1487
