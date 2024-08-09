# Results

## Test environment

NGINX Plus: true

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
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 737.955µs
Latencies     [min, mean, 50, 90, 95, 99, max]  545.611µs, 757.197µs, 734.577µs, 845.22µs, 887.95µs, 1.018ms, 12.663ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 826.722µs
Latencies     [min, mean, 50, 90, 95, 99, max]  576.284µs, 794.218µs, 779.926µs, 898.695µs, 945.876µs, 1.059ms, 11.092ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.97
Duration      [total, attack, wait]             30s, 29.999s, 770.831µs
Latencies     [min, mean, 50, 90, 95, 99, max]  558.235µs, 780.085µs, 765.963µs, 884.549µs, 927.74µs, 1.04ms, 10.613ms
Bytes In      [total, mean]                     5009833, 167.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 770.31µs
Latencies     [min, mean, 50, 90, 95, 99, max]  564.785µs, 784.552µs, 767.705µs, 886.201µs, 931.805µs, 1.046ms, 14.459ms
Bytes In      [total, mean]                     4680000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 807.239µs
Latencies     [min, mean, 50, 90, 95, 99, max]  571.22µs, 768.337µs, 753.514µs, 869.559µs, 914.12µs, 1.031ms, 9.796ms
Bytes In      [total, mean]                     4680000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
