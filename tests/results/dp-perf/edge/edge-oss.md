# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 2ed7d4ae2f827623074c40653ac821b61ae72b63
- Date: 2024-08-08T21:29:44Z
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
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 582.398µs
Latencies     [min, mean, 50, 90, 95, 99, max]  497.713µs, 646.508µs, 636.996µs, 718.685µs, 749.8µs, 837.355µs, 9.113ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 651.43µs
Latencies     [min, mean, 50, 90, 95, 99, max]  532.944µs, 671.977µs, 657.662µs, 743.313µs, 774.07µs, 862.526µs, 13.558ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.07, 1000.05
Duration      [total, attack, wait]             29.999s, 29.998s, 619.624µs
Latencies     [min, mean, 50, 90, 95, 99, max]  525.422µs, 671.622µs, 658.493µs, 749.338µs, 782.902µs, 870.726µs, 10.618ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 639.035µs
Latencies     [min, mean, 50, 90, 95, 99, max]  499.212µs, 665.644µs, 653.81µs, 739.852µs, 772.841µs, 860.586µs, 12.638ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 626.728µs
Latencies     [min, mean, 50, 90, 95, 99, max]  511.888µs, 661.603µs, 651.396µs, 730.352µs, 761.469µs, 839.857µs, 10.937ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
