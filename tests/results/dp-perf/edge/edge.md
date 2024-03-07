# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 3
- k8s version: v1.27.8-gke.1067004
- vCPUs per node: 2
- RAM per node: 4022900Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: e2-medium

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 780.444µs
Latencies     [min, mean, 50, 90, 95, 99, max]  486.416µs, 4.608ms, 1.016ms, 2.896ms, 14.596ms, 107.439ms, 230.577ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.95
Duration      [total, attack, wait]             30.002s, 30s, 1.697ms
Latencies     [min, mean, 50, 90, 95, 99, max]  581.204µs, 2.112ms, 1.111ms, 2.526ms, 5.708ms, 23.07ms, 159.71ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 793.546µs
Latencies     [min, mean, 50, 90, 95, 99, max]  543.817µs, 4.5ms, 1.139ms, 4.346ms, 16.532ms, 98.596ms, 204.682ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 1.142ms
Latencies     [min, mean, 50, 90, 95, 99, max]  553.659µs, 7.934ms, 1.106ms, 11.23ms, 43.338ms, 156.016ms, 298.312ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.10, 1000.05
Duration      [total, attack, wait]             29.998s, 29.997s, 1.55ms
Latencies     [min, mean, 50, 90, 95, 99, max]  564.824µs, 5.549ms, 1.168ms, 10.506ms, 28.084ms, 92.495ms, 148.078ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
