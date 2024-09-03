# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855004Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 584.105µs
Latencies     [min, mean, 50, 90, 95, 99, max]  467.184µs, 587.86µs, 579.209µs, 641.536µs, 663.75µs, 722.651µs, 6.479ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 631.406µs
Latencies     [min, mean, 50, 90, 95, 99, max]  486.621µs, 610.947µs, 602.417µs, 667.172µs, 692.688µs, 755.451µs, 8.14ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 588.721µs
Latencies     [min, mean, 50, 90, 95, 99, max]  502.229µs, 627.365µs, 618.853µs, 681.052µs, 706.527µs, 768.011µs, 11.029ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 558.082µs
Latencies     [min, mean, 50, 90, 95, 99, max]  489.772µs, 609.939µs, 601.244µs, 664.959µs, 689.924µs, 750.118µs, 6.527ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 541.556µs
Latencies     [min, mean, 50, 90, 95, 99, max]  493.629µs, 611.91µs, 602.091µs, 667.353µs, 694.386µs, 757.909µs, 9.034ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
