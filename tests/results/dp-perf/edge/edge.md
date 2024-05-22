# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 12
- k8s version: v1.28.8-gke.1095000
- vCPUs per node: 16
- RAM per node: 65855088Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 888.7µs
Latencies     [min, mean, 50, 90, 95, 99, max]  676.42µs, 996.68µs, 979.882µs, 1.129ms, 1.179ms, 1.32ms, 12.834ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 933.208µs
Latencies     [min, mean, 50, 90, 95, 99, max]  699.653µs, 1.01ms, 985.029µs, 1.141ms, 1.2ms, 1.341ms, 30.541ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 948.032µs
Latencies     [min, mean, 50, 90, 95, 99, max]  671.563µs, 1.037ms, 999.716µs, 1.16ms, 1.216ms, 1.438ms, 13.786ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 934.856µs
Latencies     [min, mean, 50, 90, 95, 99, max]  716.558µs, 1.04ms, 1.015ms, 1.191ms, 1.251ms, 1.425ms, 21.851ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 855.928µs
Latencies     [min, mean, 50, 90, 95, 99, max]  714.636µs, 1.032ms, 1.01ms, 1.18ms, 1.236ms, 1.394ms, 13.232ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
