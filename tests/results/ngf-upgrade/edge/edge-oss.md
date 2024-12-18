# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853980Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 1.068ms
Latencies     [min, mean, 50, 90, 95, 99, max]  665.949µs, 967.617µs, 938.253µs, 1.101ms, 1.166ms, 1.373ms, 14.936ms
Bytes In      [total, mean]                     936000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 880.619µs
Latencies     [min, mean, 50, 90, 95, 99, max]  473.276µs, 886.771µs, 867.111µs, 1.036ms, 1.094ms, 1.28ms, 12.061ms
Bytes In      [total, mean]                     969991, 161.67
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)
