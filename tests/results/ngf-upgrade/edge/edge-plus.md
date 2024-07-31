# Results

## Test environment

NGINX Plus: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1038001
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 867.428µs
Latencies     [min, mean, 50, 90, 95, 99, max]  653.414µs, 901.844µs, 879.015µs, 1.006ms, 1.051ms, 1.18ms, 12.8ms
Bytes In      [total, mean]                     919993, 153.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 839.172µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.992µs, 816.859µs, 820.308µs, 942.646µs, 983.206µs, 1.111ms, 11.939ms
Bytes In      [total, mean]                     960000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)
