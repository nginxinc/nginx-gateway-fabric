# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: e7d217a8f01fb3c8fc4507ef6f0e7feead667f20
- Date: 2024-11-14T18:42:55Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 772.721µs
Latencies     [min, mean, 50, 90, 95, 99, max]  596.14µs, 835.746µs, 798.892µs, 926.941µs, 975.553µs, 1.151ms, 14.047ms
Bytes In      [total, mean]                     956060, 159.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 955.654µs
Latencies     [min, mean, 50, 90, 95, 99, max]  630.575µs, 856.057µs, 818.559µs, 939.26µs, 986.329µs, 1.177ms, 14.025ms
Bytes In      [total, mean]                     918000, 153.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)
