# Results

## Test environment

NGINX Plus: false

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

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 997.51
Duration      [total, attack, wait]             30s, 29.999s, 727.553µs
Latencies     [min, mean, 50, 90, 95, 99, max]  345.374µs, 731.555µs, 712.731µs, 813.3µs, 851.098µs, 954.796µs, 23.668ms
Bytes In      [total, mean]                     4772325, 159.08
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.75%
Status Codes  [code:count]                      200:29925  503:75  
Error Set:
503 Service Temporarily Unavailable
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 773.602µs
Latencies     [min, mean, 50, 90, 95, 99, max]  538.108µs, 760.895µs, 747.297µs, 847.355µs, 884.042µs, 994.549µs, 11.071ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 852.223µs
Latencies     [min, mean, 50, 90, 95, 99, max]  557.936µs, 764.327µs, 752.802µs, 858.304µs, 897.981µs, 1.007ms, 7.479ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 732.366µs
Latencies     [min, mean, 50, 90, 95, 99, max]  556.067µs, 763.809µs, 749.494µs, 854.451µs, 892.688µs, 1.004ms, 11.467ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 626.217µs
Latencies     [min, mean, 50, 90, 95, 99, max]  559.252µs, 764.401µs, 747.653µs, 848.203µs, 885.688µs, 996.553µs, 14.815ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
