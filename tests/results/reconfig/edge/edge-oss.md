# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: bf8ea47203eb4695af0d359243c73de2d1badbbf
- Date: 2024-09-13T20:33:11Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1639000
- vCPUs per node: 16
- RAM per node: 65853968Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test 1: Resources exist before startup - NumResources 30

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

- Event Batch Total: 7
- Event Batch Processing Average Time: 39ms
- Event Batch Processing distribution:
	- 500ms: 7
	- 1000ms: 7
	- 5000ms: 7
	- 10000ms: 7
	- 30000ms: 7
	- +Infms: 7

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

- Event Batch Total: 338
- Event Batch Processing Average Time: 23ms
- Event Batch Processing distribution:
	- 500ms: 338
	- 1000ms: 338
	- 5000ms: 338
	- 10000ms: 338
	- 30000ms: 338
	- +Infms: 338

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 44s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 342
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 342
	- 1000ms: 342
	- 5000ms: 342
	- 10000ms: 342
	- 30000ms: 342
	- +Infms: 342

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

- TimeToReadyTotal: 1s
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

- Event Batch Total: 1554
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1554
	- 1000ms: 1554
	- 5000ms: 1554
	- 10000ms: 1554
	- 30000ms: 1554
	- +Infms: 1554
