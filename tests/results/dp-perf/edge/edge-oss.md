# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
- vCPUs per node: 16
- RAM per node: 65853964Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 999.50
Duration      [total, attack, wait]             30s, 29.999s, 779.26µs
Latencies     [min, mean, 50, 90, 95, 99, max]  401.546µs, 729.594µs, 696.171µs, 808.825µs, 853.548µs, 993.26µs, 12.557ms
Bytes In      [total, mean]                     4769865, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.95%
Status Codes  [code:count]                      200:29985  502:15  
Error Set:
502 Bad Gateway
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 707.055µs
Latencies     [min, mean, 50, 90, 95, 99, max]  535.388µs, 765.759µs, 741.731µs, 856.44µs, 900.22µs, 1.028ms, 12.624ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 825.751µs
Latencies     [min, mean, 50, 90, 95, 99, max]  562.053µs, 766.667µs, 744.972µs, 859.534µs, 901.417µs, 1.02ms, 11.739ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 794.715µs
Latencies     [min, mean, 50, 90, 95, 99, max]  550.188µs, 749.652µs, 735.445µs, 851.278µs, 896.326µs, 1.024ms, 5.942ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 776.178µs
Latencies     [min, mean, 50, 90, 95, 99, max]  555.668µs, 750.151µs, 736.996µs, 847.199µs, 890.538µs, 1.003ms, 9.982ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
