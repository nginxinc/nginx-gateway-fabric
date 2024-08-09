# Results

## Test environment

NGINX Plus: false

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
Duration      [total, attack, wait]             59.992s, 59.991s, 903.286µs
Latencies     [min, mean, 50, 90, 95, 99, max]  663.853µs, 993.737µs, 947.332µs, 1.156ms, 1.236ms, 1.5ms, 13.09ms
Bytes In      [total, mean]                     912000, 152.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 966.094µs
Latencies     [min, mean, 50, 90, 95, 99, max]  621.594µs, 937.441µs, 902.062µs, 1.087ms, 1.153ms, 1.345ms, 12.671ms
Bytes In      [total, mean]                     954000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)
