# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: bf8ea47203eb4695af0d359243c73de2d1badbbf
- Date: 2024-09-13T20:33:11Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1639000
- vCPUs per node: 16
- RAM per node: 65853960Ki
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

- TimeToReadyTotal: 4s
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

- TimeToReadyTotal: 7s
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

- Event Batch Total: 335
- Event Batch Processing Average Time: 23ms
- Event Batch Processing distribution:
	- 500ms: 335
	- 1000ms: 335
	- 5000ms: 335
	- 10000ms: 335
	- 30000ms: 335
	- +Infms: 335

## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
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

- Event Batch Total: 1691
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1691
	- 1000ms: 1691
	- 5000ms: 1691
	- 10000ms: 1691
	- 30000ms: 1691
	- +Infms: 1691

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

- Event Batch Total: 1556
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1556
	- 1000ms: 1556
	- 5000ms: 1556
	- 10000ms: 1556
	- 30000ms: 1556
	- +Infms: 1556
