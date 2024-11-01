# Results

## Test environment

NGINX Plus: true

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

- TimeToReadyTotal: 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 114ms
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
- NGINX Reloads: 61
- NGINX Reload Average Time: 128ms
- Reload distribution:
	- 500ms: 61
	- 1000ms: 61
	- 5000ms: 61
	- 10000ms: 61
	- 30000ms: 61
	- +Infms: 61

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
- NGINX Reloads: 341
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 341
	- 1000ms: 341
	- 5000ms: 341
	- 10000ms: 341
	- 30000ms: 341
	- +Infms: 341

### Event Batch Processing

- Event Batch Total: 1693
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500ms: 1693
	- 1000ms: 1693
	- 5000ms: 1693
	- 10000ms: 1693
	- 30000ms: 1693
	- +Infms: 1693

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

- Event Batch Total: 303
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 303
	- 1000ms: 303
	- 5000ms: 303
	- 10000ms: 303
	- 30000ms: 303
	- +Infms: 303

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

- Event Batch Total: 1556
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500ms: 1556
	- 1000ms: 1556
	- 5000ms: 1556
	- 10000ms: 1556
	- 30000ms: 1556
	- +Infms: 1556
