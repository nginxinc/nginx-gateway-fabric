# Results

## Test environment

NGINX Plus: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.4-gke.1043004
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.99s, 1.631ms
Latencies     [min, mean, 50, 90, 95, 99, max]  908.3µs, 1.36ms, 1.329ms, 1.511ms, 1.57ms, 1.795ms, 14.259ms
Bytes In      [total, mean]                     955989, 159.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.99s, 1.671ms
Latencies     [min, mean, 50, 90, 95, 99, max]  968.82µs, 1.388ms, 1.338ms, 1.558ms, 1.632ms, 1.973ms, 14.346ms
Bytes In      [total, mean]                     912000, 152.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)
