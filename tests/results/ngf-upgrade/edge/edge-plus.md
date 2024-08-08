# Results

## Test environment

NGINX Plus: true

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
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 872.066µs
Latencies     [min, mean, 50, 90, 95, 99, max]  645.151µs, 908.553µs, 868.227µs, 1.028ms, 1.081ms, 1.296ms, 12.693ms
Bytes In      [total, mean]                     906000, 151.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 800.464µs
Latencies     [min, mean, 50, 90, 95, 99, max]  634.294µs, 877.396µs, 842.202µs, 972.867µs, 1.028ms, 1.249ms, 12.692ms
Bytes In      [total, mean]                     949941, 158.32
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)
