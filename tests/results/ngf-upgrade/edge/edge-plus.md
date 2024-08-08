# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 809c0838e2f2658c3c4cd48325ffb0bc5a92a002
- Date: 2024-08-08T18:03:35Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1254000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 913.832µs
Latencies     [min, mean, 50, 90, 95, 99, max]  447.468µs, 877.575µs, 872.625µs, 1.029ms, 1.081ms, 1.224ms, 7.517ms
Bytes In      [total, mean]                     920041, 153.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 854.57µs
Latencies     [min, mean, 50, 90, 95, 99, max]  436.603µs, 833.46µs, 843.163µs, 992.309µs, 1.046ms, 1.198ms, 5.381ms
Bytes In      [total, mean]                     962034, 160.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)
