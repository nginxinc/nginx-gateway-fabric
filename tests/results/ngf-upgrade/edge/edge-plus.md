# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853984Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 86.13
Duration      [total, attack, wait]             59.993s, 59.99s, 2.942ms
Latencies     [min, mean, 50, 90, 95, 99, max]  471.41µs, 1.25ms, 949.477µs, 2.838ms, 3.054ms, 3.539ms, 12.939ms
Bytes In      [total, mean]                     955693, 159.28
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           86.12%
Status Codes  [code:count]                      200:5167  503:833  
Error Set:
503 Service Temporarily Unavailable
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 86.13
Duration      [total, attack, wait]             59.991s, 59.99s, 739.811µs
Latencies     [min, mean, 50, 90, 95, 99, max]  607.774µs, 941.816µs, 923.315µs, 1.045ms, 1.092ms, 1.279ms, 12.42ms
Bytes In      [total, mean]                     984990, 164.16
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           86.12%
Status Codes  [code:count]                      200:5167  503:833  
Error Set:
503 Service Temporarily Unavailable
```

![http-plus.png](http-plus.png)
