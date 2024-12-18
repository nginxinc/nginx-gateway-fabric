# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853984Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 634.387µs
Latencies     [min, mean, 50, 90, 95, 99, max]  556.824µs, 790.374µs, 774.114µs, 899.825µs, 947.466µs, 1.076ms, 11.134ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 821.625µs
Latencies     [min, mean, 50, 90, 95, 99, max]  580.185µs, 825.124µs, 807.498µs, 942.495µs, 992.13µs, 1.115ms, 11.269ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 831.445µs
Latencies     [min, mean, 50, 90, 95, 99, max]  586.372µs, 839.603µs, 820.061µs, 945.712µs, 994.204µs, 1.128ms, 12.904ms
Bytes In      [total, mean]                     5100000, 170.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 927.681µs
Latencies     [min, mean, 50, 90, 95, 99, max]  579.176µs, 821.601µs, 804.632µs, 936.14µs, 986.749µs, 1.125ms, 8.856ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 903.577µs
Latencies     [min, mean, 50, 90, 95, 99, max]  577.936µs, 828.521µs, 812.691µs, 947.089µs, 997.055µs, 1.122ms, 10.048ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
