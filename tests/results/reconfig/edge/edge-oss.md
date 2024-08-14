# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 9a85dbcc0797e31557a3731688795aa166ee0f96
- Date: 2024-08-13T21:12:05Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1326000
- vCPUs per node: 16
- RAM per node: 65855012Ki
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
- Event Batch Processing Average Time: 47ms
- Event Batch Processing distribution:
	- 500ms: 6
	- 1000ms: 6
	- 5000ms: 6
	- 10000ms: 6
	- 30000ms: 6
	- +Infms: 6


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

- Event Batch Total: 305
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 305
	- 1000ms: 305
	- 5000ms: 305
	- 10000ms: 305
	- 30000ms: 305
	- +Infms: 305


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 343
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 343
	- 1000ms: 343
	- 5000ms: 343
	- 10000ms: 343
	- 30000ms: 343
	- +Infms: 343

### Event Batch Processing

- Event Batch Total: 1543
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500ms: 1543
	- 1000ms: 1543
	- 5000ms: 1543
	- 10000ms: 1543
	- 30000ms: 1543
	- +Infms: 1543

