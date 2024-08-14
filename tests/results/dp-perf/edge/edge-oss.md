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

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.79
Duration      [total, attack, wait]             30s, 30s, 701.442µs
Latencies     [min, mean, 50, 90, 95, 99, max]  341.042µs, 737.394µs, 718.76µs, 827.697µs, 866.432µs, 962.831µs, 16.145ms
Bytes In      [total, mean]                     4769946, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.98%
Status Codes  [code:count]                      200:29994  502:6  
Error Set:
502 Bad Gateway
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 687.856µs
Latencies     [min, mean, 50, 90, 95, 99, max]  554.445µs, 759.558µs, 740.252µs, 848.701µs, 889.401µs, 997.689µs, 22.657ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 726.197µs
Latencies     [min, mean, 50, 90, 95, 99, max]  550.722µs, 769.327µs, 747.389µs, 857.742µs, 898.394µs, 1.015ms, 13.026ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 705.904µs
Latencies     [min, mean, 50, 90, 95, 99, max]  539.272µs, 750.546µs, 734.78µs, 841.798µs, 880.351µs, 979.455µs, 16.567ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 772.597µs
Latencies     [min, mean, 50, 90, 95, 99, max]  541.485µs, 752.07µs, 740.31µs, 847.409µs, 886.955µs, 981.786µs, 12.398ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
