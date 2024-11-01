# Results

## Test environment

NGINX Plus: true

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

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 840.484µs
Latencies     [min, mean, 50, 90, 95, 99, max]  655.564µs, 940.948µs, 915.293µs, 1.076ms, 1.13ms, 1.289ms, 12.465ms
Bytes In      [total, mean]                     961995, 160.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 686.428µs
Latencies     [min, mean, 50, 90, 95, 99, max]  516.584µs, 954.477µs, 934.142µs, 1.096ms, 1.157ms, 1.332ms, 15.112ms
Bytes In      [total, mean]                     924000, 154.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)
