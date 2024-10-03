# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
- vCPUs per node: 16
- RAM per node: 65853964Ki
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

- TimeToReadyTotal: 3s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 76ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 5
- Event Batch Processing Average Time: 94ms
- Event Batch Processing distribution:
	- 500ms: 5
	- 1000ms: 5
	- 5000ms: 5
	- 10000ms: 5
	- 30000ms: 5
	- +Infms: 5

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

- Event Batch Total: 339
- Event Batch Processing Average Time: 23ms
- Event Batch Processing distribution:
	- 500ms: 339
	- 1000ms: 339
	- 5000ms: 339
	- 10000ms: 339
	- 30000ms: 339
	- +Infms: 339

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 339
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 339
	- 1000ms: 339
	- 5000ms: 339
	- 10000ms: 339
	- 30000ms: 339
	- +Infms: 339

### Event Batch Processing

- Event Batch Total: 1692
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1692
	- 1000ms: 1692
	- 5000ms: 1692
	- 10000ms: 1692
	- 30000ms: 1692
	- +Infms: 1692

## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 64
- NGINX Reload Average Time: 126ms
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
- NGINX Reloads: 343
- NGINX Reload Average Time: 125ms
- Reload distribution:
	- 500ms: 343
	- 1000ms: 343
	- 5000ms: 343
	- 10000ms: 343
	- 30000ms: 343
	- +Infms: 343

### Event Batch Processing

- Event Batch Total: 1545
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1545
	- 1000ms: 1545
	- 5000ms: 1545
	- 10000ms: 1545
	- 30000ms: 1545
	- +Infms: 1545
