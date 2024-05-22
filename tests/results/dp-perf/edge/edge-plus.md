# Results

## Test environment

NGINX Plus: true

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
Duration      [total, attack, wait]             30s, 29.999s, 901.249µs
Latencies     [min, mean, 50, 90, 95, 99, max]  726.198µs, 1.027ms, 1.002ms, 1.141ms, 1.187ms, 1.32ms, 21.478ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 1.126ms
Latencies     [min, mean, 50, 90, 95, 99, max]  754.249µs, 1.053ms, 1.035ms, 1.184ms, 1.235ms, 1.383ms, 11.298ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.98
Duration      [total, attack, wait]             30.001s, 29.999s, 1.228ms
Latencies     [min, mean, 50, 90, 95, 99, max]  746.179µs, 1.033ms, 1.018ms, 1.163ms, 1.211ms, 1.341ms, 11.734ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.97
Duration      [total, attack, wait]             30.001s, 29.999s, 1.473ms
Latencies     [min, mean, 50, 90, 95, 99, max]  737.645µs, 1.054ms, 1.031ms, 1.188ms, 1.254ms, 1.587ms, 11.03ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 1ms
Latencies     [min, mean, 50, 90, 95, 99, max]  752.767µs, 1.027ms, 1.006ms, 1.143ms, 1.193ms, 1.305ms, 21.596ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
