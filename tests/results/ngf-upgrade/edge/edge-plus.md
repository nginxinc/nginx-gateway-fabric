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

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 825.711µs
Latencies     [min, mean, 50, 90, 95, 99, max]  629.57µs, 918.636µs, 874.76µs, 1.036ms, 1.098ms, 1.312ms, 12.532ms
Bytes In      [total, mean]                     912000, 152.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 577.325µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.641µs, 883.927µs, 851.926µs, 1.006ms, 1.068ms, 1.317ms, 12.988ms
Bytes In      [total, mean]                     955970, 159.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)
