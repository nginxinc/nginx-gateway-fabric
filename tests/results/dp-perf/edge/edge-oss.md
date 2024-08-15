# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3c029b1417c1f89f2a29aeef07f47078640e28b2
- Date: 2024-08-15T00:04:25Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1326000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 999.91
Duration      [total, attack, wait]             30s, 29.999s, 586.041µs
Latencies     [min, mean, 50, 90, 95, 99, max]  459.105µs, 652.462µs, 640.858µs, 731.955µs, 766.916µs, 868.446µs, 5.683ms
Bytes In      [total, mean]                     4829967, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.99%
Status Codes  [code:count]                      200:29997  502:3  
Error Set:
502 Bad Gateway
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 637.326µs
Latencies     [min, mean, 50, 90, 95, 99, max]  474.92µs, 694.704µs, 678.367µs, 786.23µs, 826.769µs, 934.323µs, 12.252ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 30s, 676.679µs
Latencies     [min, mean, 50, 90, 95, 99, max]  525.067µs, 708.22µs, 693.208µs, 804.086µs, 849.031µs, 958.899µs, 10.449ms
Bytes In      [total, mean]                     5100000, 170.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 637.156µs
Latencies     [min, mean, 50, 90, 95, 99, max]  512.28µs, 691.498µs, 675.513µs, 780.523µs, 817.834µs, 910.375µs, 9.619ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 773.157µs
Latencies     [min, mean, 50, 90, 95, 99, max]  503.29µs, 669.335µs, 655.455µs, 754.428µs, 792.559µs, 892.505µs, 10.652ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
