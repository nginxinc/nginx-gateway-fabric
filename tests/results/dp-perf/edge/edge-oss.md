# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: e7d217a8f01fb3c8fc4507ef6f0e7feead667f20
- Date: 2024-11-14T18:42:55Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 546.336µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.328µs, 652.455µs, 635.612µs, 733.819µs, 775.191µs, 898.404µs, 12.115ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 726.771µs
Latencies     [min, mean, 50, 90, 95, 99, max]  516.729µs, 667.447µs, 651.234µs, 748.002µs, 789.132µs, 912.936µs, 12.01ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         29999, 1000.02, 1000.00
Duration      [total, attack, wait]             29.999s, 29.998s, 775.462µs
Latencies     [min, mean, 50, 90, 95, 99, max]  509.103µs, 675.843µs, 660.979µs, 753.464µs, 790.594µs, 915.896µs, 10.924ms
Bytes In      [total, mean]                     5099830, 170.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             30s, 29.999s, 579.805µs
Latencies     [min, mean, 50, 90, 95, 99, max]  495.375µs, 663.275µs, 645.382µs, 747.059µs, 791.431µs, 925.516µs, 10.063ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 604.093µs
Latencies     [min, mean, 50, 90, 95, 99, max]  499.363µs, 654.309µs, 639.452µs, 737.027µs, 777.872µs, 904.014µs, 8.053ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
