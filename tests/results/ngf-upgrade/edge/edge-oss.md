# Results

## Test environment

NGINX Plus: false

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

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 856.511µs
Latencies     [min, mean, 50, 90, 95, 99, max]  636.875µs, 900.917µs, 859.919µs, 1.023ms, 1.093ms, 1.294ms, 12.65ms
Bytes In      [total, mean]                     954000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 878.572µs
Latencies     [min, mean, 50, 90, 95, 99, max]  704.008µs, 1.019ms, 977.8µs, 1.176ms, 1.249ms, 1.45ms, 18.529ms
Bytes In      [total, mean]                     912000, 152.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)
