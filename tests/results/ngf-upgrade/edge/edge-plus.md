# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 3a08fdafadfe0fb4a9c25679da1a1fcd6b181474
- Date: 2024-10-15T13:45:52Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1014001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 913.585µs
Latencies     [min, mean, 50, 90, 95, 99, max]  660.737µs, 910.81µs, 889.797µs, 1.034ms, 1.093ms, 1.262ms, 14.865ms
Bytes In      [total, mean]                     968001, 161.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 939.801µs
Latencies     [min, mean, 50, 90, 95, 99, max]  484.682µs, 898.425µs, 890.482µs, 1.025ms, 1.079ms, 1.209ms, 13.209ms
Bytes In      [total, mean]                     932025, 155.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)
