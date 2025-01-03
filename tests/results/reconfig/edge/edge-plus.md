# Results

## Test environment

NGINX Plus: true

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

- TimeToReadyTotal: 5s
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
- Event Batch Processing Average Time: 57ms
- Event Batch Processing distribution:
	- 500.0ms: 6
	- 1000.0ms: 6
	- 5000.0ms: 6
	- 10000.0ms: 6
	- 30000.0ms: 6
	- +Infms: 6

### NGINX Error Logs


## Test 1: Resources exist before startup - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 5s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 2
- NGINX Reload Average Time: 139ms
- Reload distribution:
	- 500.0ms: 2
	- 1000.0ms: 2
	- 5000.0ms: 2
	- 10000.0ms: 2
	- 30000.0ms: 2
	- +Infms: 2

### Event Batch Processing

- Event Batch Total: 6
- Event Batch Processing Average Time: 65ms
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

- TimeToReadyTotal: 7s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 45
- NGINX Reload Average Time: 155ms
- Reload distribution:
	- 500.0ms: 45
	- 1000.0ms: 45
	- 5000.0ms: 45
	- 10000.0ms: 45
	- 30000.0ms: 45
	- +Infms: 45

### Event Batch Processing

- Event Batch Total: 321
- Event Batch Processing Average Time: 25ms
- Event Batch Processing distribution:
	- 500.0ms: 321
	- 1000.0ms: 321
	- 5000.0ms: 321
	- 10000.0ms: 321
	- 30000.0ms: 321
	- +Infms: 321

### NGINX Error Logs


## Test 2: Start NGF, deploy Gateway, create many resources attached to GW - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: 43s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 245
- NGINX Reload Average Time: 152ms
- Reload distribution:
	- 500.0ms: 245
	- 1000.0ms: 245
	- 5000.0ms: 245
	- 10000.0ms: 245
	- 30000.0ms: 245
	- +Infms: 245

### Event Batch Processing

- Event Batch Total: 1600
- Event Batch Processing Average Time: 27ms
- Event Batch Processing distribution:
	- 500.0ms: 1600
	- 1000.0ms: 1600
	- 5000.0ms: 1600
	- 10000.0ms: 1600
	- 30000.0ms: 1600
	- +Infms: 1600

### NGINX Error Logs


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 30

### Reloads and Time to Ready

- TimeToReadyTotal: -4s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 25
- NGINX Reload Average Time: 149ms
- Reload distribution:
	- 500.0ms: 25
	- 1000.0ms: 25
	- 5000.0ms: 25
	- 10000.0ms: 25
	- 30000.0ms: 25
	- +Infms: 25

### Event Batch Processing

- Event Batch Total: 266
- Event Batch Processing Average Time: 18ms
- Event Batch Processing distribution:
	- 500.0ms: 266
	- 1000.0ms: 266
	- 5000.0ms: 266
	- 10000.0ms: 266
	- 30000.0ms: 266
	- +Infms: 266

### NGINX Error Logs
2025/01/03 21:27:53 [emerg] 45#45: invalid instance state file "/var/lib/nginx/state/nginx-mgmt-state"


## Test 3: Start NGF, create many resources attached to a Gateway, deploy the Gateway - NumResources 150

### Reloads and Time to Ready

- TimeToReadyTotal: < 1s
- TimeToReadyAvgSingle: < 1s
- NGINX Reloads: 231
- NGINX Reload Average Time: 151ms
- Reload distribution:
	- 500.0ms: 231
	- 1000.0ms: 231
	- 5000.0ms: 231
	- 10000.0ms: 231
	- 30000.0ms: 231
	- +Infms: 231

### Event Batch Processing

- Event Batch Total: 1447
- Event Batch Processing Average Time: 30ms
- Event Batch Processing distribution:
	- 500.0ms: 1446
	- 1000.0ms: 1447
	- 5000.0ms: 1447
	- 10000.0ms: 1447
	- 30000.0ms: 1447
	- +Infms: 1447

### NGINX Error Logs

