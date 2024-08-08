# Results

## Test environment

NGINX Plus: false

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
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 748.239µs
Latencies     [min, mean, 50, 90, 95, 99, max]  522.64µs, 754.555µs, 714.889µs, 828.966µs, 875.27µs, 1.018ms, 17.567ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 753.203µs
Latencies     [min, mean, 50, 90, 95, 99, max]  563.123µs, 768.346µs, 753.233µs, 875.083µs, 923.38µs, 1.054ms, 7.245ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 733.057µs
Latencies     [min, mean, 50, 90, 95, 99, max]  572.366µs, 773.623µs, 754.835µs, 876.485µs, 924.854µs, 1.047ms, 13.662ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30s, 30s, 647.613µs
Latencies     [min, mean, 50, 90, 95, 99, max]  537.821µs, 773.652µs, 752.523µs, 881.602µs, 930.635µs, 1.065ms, 17.091ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 771.879µs
Latencies     [min, mean, 50, 90, 95, 99, max]  561.163µs, 755.997µs, 731.634µs, 848.266µs, 896.817µs, 1.025ms, 12.797ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
