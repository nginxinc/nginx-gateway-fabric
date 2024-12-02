# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 929413c15af7bee3adb32e103c9d1513a693da16
- Date: 2024-11-28T12:52:45Z
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

- TimeToReadyTotal: 4s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500.0ms: 2
	- 1000.0ms: 2
	- 5000.0ms: 2
	- 10000.0ms: 2
	- 30000.0ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 60ms
- Event Batch Processing distribution:
	- 500.0ms: 6
	- 1000.0ms: 6
	- 5000.0ms: 6
	- 10000.0ms: 6
	- 30000.0ms: 6
	- +Infms: 6

## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 2s
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
- Event Batch Processing Average Time: 62ms
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
- NGINX Reloads: 44
- NGINX Reload Average Time: 153ms
- Reload distribution:
	- 500.0ms: 44
	- 1000.0ms: 44
	- 5000.0ms: 44
	- 10000.0ms: 44
	- 30000.0ms: 44
	- +Infms: 44

### Event Batch Processing

- Event Batch Total: 319
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500.0ms: 319
	- 1000.0ms: 319
	- 5000.0ms: 319
	- 10000.0ms: 319
	- 30000.0ms: 319
	- +Infms: 319

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 44s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 235
- NGINX Reload Average Time: 152ms
- Reload distribution:
	- 500.0ms: 235
	- 1000.0ms: 235
	- 5000.0ms: 235
	- 10000.0ms: 235
	- 30000.0ms: 235
	- +Infms: 235

### Event Batch Processing

- Event Batch Total: 1591
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500.0ms: 1591
	- 1000.0ms: 1591
	- 5000.0ms: 1591
	- 10000.0ms: 1591
	- 30000.0ms: 1591
	- +Infms: 1591

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 44
- NGINX Reload Average Time: 150ms
- Reload distribution:
	- 500.0ms: 44
	- 1000.0ms: 44
	- 5000.0ms: 44
	- 10000.0ms: 44
	- 30000.0ms: 44
	- +Infms: 44

### Event Batch Processing

- Event Batch Total: 287
- Event Batch Processing Average Time: 29ms
- Event Batch Processing distribution:
	- 500.0ms: 287
	- 1000.0ms: 287
	- 5000.0ms: 287
	- 10000.0ms: 287
	- 30000.0ms: 287
	- +Infms: 287

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 217
- NGINX Reload Average Time: 151ms
- Reload distribution:
	- 500.0ms: 217
	- 1000.0ms: 217
	- 5000.0ms: 217
	- 10000.0ms: 217
	- 30000.0ms: 217
	- +Infms: 217

### Event Batch Processing

- Event Batch Total: 1432
- Event Batch Processing Average Time: 30ms
- Event Batch Processing distribution:
	- 500.0ms: 1431
	- 1000.0ms: 1432
	- 5000.0ms: 1432
	- 10000.0ms: 1432
	- 30000.0ms: 1432
	- +Infms: 1432
