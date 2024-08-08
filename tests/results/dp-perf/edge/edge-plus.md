# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 809c0838e2f2658c3c4cd48325ffb0bc5a92a002
- Date: 2024-08-08T18:03:35Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1254000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 660.838µs
Latencies     [min, mean, 50, 90, 95, 99, max]  482.186µs, 694.186µs, 681.526µs, 779.766µs, 817.831µs, 925.007µs, 11.331ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 692.691µs
Latencies     [min, mean, 50, 90, 95, 99, max]  528.62µs, 727.916µs, 703.627µs, 801.065µs, 839.815µs, 955.774µs, 21.995ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 754.34µs
Latencies     [min, mean, 50, 90, 95, 99, max]  545.277µs, 718.685µs, 706.698µs, 805.322µs, 844.749µs, 961.819µs, 6.804ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 685.126µs
Latencies     [min, mean, 50, 90, 95, 99, max]  531.98µs, 720.656µs, 707.127µs, 814.111µs, 857.45µs, 980.744µs, 8.717ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 718.275µs
Latencies     [min, mean, 50, 90, 95, 99, max]  546.498µs, 717.715µs, 704.964µs, 805.675µs, 843.967µs, 952.086µs, 14.292ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
