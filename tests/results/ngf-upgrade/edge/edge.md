# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 3
- k8s version: v1.27.8-gke.1067004
- vCPUs per node: 2
- RAM per node: 4022900Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: e2-medium

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 1.183ms
Latencies     [min, mean, 50, 90, 95, 99, max]  838.275µs, 1.866ms, 1.718ms, 2.294ms, 2.601ms, 5.354ms, 40.474ms
Bytes In      [total, mean]                     955975, 159.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http.png](http.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.99s, 1.519ms
Latencies     [min, mean, 50, 90, 95, 99, max]  952.224µs, 2.056ms, 1.802ms, 2.446ms, 3.076ms, 8.224ms, 41.937ms
Bytes In      [total, mean]                     916005, 152.67
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https.png](https.png)
