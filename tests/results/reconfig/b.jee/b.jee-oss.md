# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 19f98ab76481e5de0ff8b9ca4ab618c7995cb90d
- Date: 2024-08-09T20:56:49Z
- Dirty: true

GKE Cluster:

- Node count: 4
- k8s version: v1.29.6-gke.1326000
- vCPUs per node: 2
- RAM per node: 4019160Ki
- Max pods per node: 110
- Zone: us-central1-c
- Instance Type: e2-medium

## Test 1: Resources exist before startup - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 2s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 140ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 76ms
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
- NGINX Reload Average Time: 127ms
- Reload distribution:
	- 500ms: 2
	- 1000ms: 2
	- 5000ms: 2
	- 10000ms: 2
	- 30000ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 5
- Event Batch Processing Average Time: 74ms
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
- NGINX Reloads: 59
- NGINX Reload Average Time: 131ms
- Reload distribution:
	- 500ms: 59
	- 1000ms: 59
	- 5000ms: 59
	- 10000ms: 59
	- 30000ms: 59
	- +Infms: 59

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
- NGINX Reloads: 310
- NGINX Reload Average Time: 135ms
- Reload distribution:
	- 500ms: 310
	- 1000ms: 310
	- 5000ms: 310
	- 10000ms: 310
	- 30000ms: 310
	- +Infms: 310

### Event Batch Processing

- Event Batch Total: 1645
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 1645
	- 1000ms: 1645
	- 5000ms: 1645
	- 10000ms: 1645
	- 30000ms: 1645
	- +Infms: 1645


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 63
- NGINX Reload Average Time: 126ms
- Reload distribution:
	- 500ms: 63
	- 1000ms: 63
	- 5000ms: 63
	- 10000ms: 63
	- 30000ms: 63
	- +Infms: 63

### Event Batch Processing

- Event Batch Total: 309
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 309
	- 1000ms: 309
	- 5000ms: 309
	- 10000ms: 309
	- 30000ms: 309
	- +Infms: 309


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 326
- NGINX Reload Average Time: 131ms
- Reload distribution:
	- 500ms: 326
	- 1000ms: 326
	- 5000ms: 326
	- 10000ms: 326
	- 30000ms: 326
	- +Infms: 326

### Event Batch Processing

- Event Batch Total: 1632
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500ms: 1632
	- 1000ms: 1632
	- 5000ms: 1632
	- 10000ms: 1632
	- 30000ms: 1632
	- +Infms: 1632
