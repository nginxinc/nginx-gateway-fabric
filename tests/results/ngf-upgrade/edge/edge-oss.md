# Results

## Test environment

NGINX Plus: false

 NGINX Gateway Fabric:

- Commit: unknown
- Date: unknown
- Dirty: unknown

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
Duration      [total, attack, wait]             59.991s, 59.99s, 955.966µs
Latencies     [min, mean, 50, 90, 95, 99, max]  713.881µs, 992.909µs, 945.057µs, 1.091ms, 1.145ms, 1.382ms, 14.396ms
Bytes In      [total, mean]                     922013, 153.67
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.991s, 59.99s, 953.507µs
Latencies     [min, mean, 50, 90, 95, 99, max]  664.913µs, 975.361µs, 919.186µs, 1.063ms, 1.114ms, 1.436ms, 13.806ms
Bytes In      [total, mean]                     960000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)
