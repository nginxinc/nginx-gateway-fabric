# Results

## Test environment

NGINX Plus: true

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
- Event Batch Processing Average Time: 41ms
- Event Batch Processing distribution:
	- 500ms: 7
	- 1000ms: 7
	- 5000ms: 7
	- 10000ms: 7
	- 30000ms: 7
	- +Infms: 7


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

- Event Batch Total: 1697
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1697
	- 1000ms: 1697
	- 5000ms: 1697
	- 10000ms: 1697
	- 30000ms: 1697
	- +Infms: 1697


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
- NGINX Reloads: 346
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 346
	- 1000ms: 346
	- 5000ms: 346
	- 10000ms: 346
	- 30000ms: 346
	- +Infms: 346

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

