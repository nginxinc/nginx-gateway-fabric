# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 9a47b6d33e6c5e0568af0bfc367e5f1190354b05
- Date: 2025-01-03T19:08:47Z
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

- TimeToReadyTotal: 3s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 127ms
- Reload distribution:
	- 500.0ms: 2
	- 1000.0ms: 2
	- 5000.0ms: 2
	- 10000.0ms: 2
	- 30000.0ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 5
- Event Batch Processing Average Time: 59ms
- Event Batch Processing distribution:
	- 500.0ms: 5
	- 1000.0ms: 5
	- 5000.0ms: 5
	- 10000.0ms: 5
	- 30000.0ms: 5
	- +Infms: 5

### NGINX Error Logs


## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 3s
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

### NGINX Error Logs


## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: 8s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 53
- NGINX Reload Average Time: 150ms
- Reload distribution:
	- 500.0ms: 53
	- 1000.0ms: 53
	- 5000.0ms: 53
	- 10000.0ms: 53
	- 30000.0ms: 53
	- +Infms: 53

### Event Batch Processing

- Event Batch Total: 329
- Event Batch Processing Average Time: 24ms
- Event Batch Processing distribution:
	- 500.0ms: 329
	- 1000.0ms: 329
	- 5000.0ms: 329
	- 10000.0ms: 329
	- 30000.0ms: 329
	- +Infms: 329

### NGINX Error Logs


## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 287
- NGINX Reload Average Time: 151ms
- Reload distribution:
	- 500.0ms: 287
	- 1000.0ms: 287
	- 5000.0ms: 287
	- 10000.0ms: 287
	- 30000.0ms: 287
	- +Infms: 287

### Event Batch Processing

- Event Batch Total: 1640
- Event Batch Processing Average Time: 26ms
- Event Batch Processing distribution:
	- 500.0ms: 1640
	- 1000.0ms: 1640
	- 5000.0ms: 1640
	- 10000.0ms: 1640
	- 30000.0ms: 1640
	- +Infms: 1640

### NGINX Error Logs


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 58
- NGINX Reload Average Time: 139ms
- Reload distribution:
	- 500.0ms: 58
	- 1000.0ms: 58
	- 5000.0ms: 58
	- 10000.0ms: 58
	- 30000.0ms: 58
	- +Infms: 58

### Event Batch Processing

- Event Batch Total: 300
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500.0ms: 300
	- 1000.0ms: 300
	- 5000.0ms: 300
	- 10000.0ms: 300
	- 30000.0ms: 300
	- +Infms: 300

### NGINX Error Logs


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 324
- NGINX Reload Average Time: 134ms
- Reload distribution:
	- 500.0ms: 324
	- 1000.0ms: 324
	- 5000.0ms: 324
	- 10000.0ms: 324
	- 30000.0ms: 324
	- +Infms: 324

### Event Batch Processing

- Event Batch Total: 1520
- Event Batch Processing Average Time: 28ms
- Event Batch Processing distribution:
	- 500.0ms: 1520
	- 1000.0ms: 1520
	- 5000.0ms: 1520
	- 10000.0ms: 1520
	- 30000.0ms: 1520
	- +Infms: 1520

### NGINX Error Logs

