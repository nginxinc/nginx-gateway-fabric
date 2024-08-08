# Results

## Test environment

NGINX Plus: true

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
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 790.051µs
Latencies     [min, mean, 50, 90, 95, 99, max]  539.113µs, 939.415µs, 931.18µs, 1.073ms, 1.122ms, 1.273ms, 7.441ms
Bytes In      [total, mean]                     920022, 153.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 1.055ms
Latencies     [min, mean, 50, 90, 95, 99, max]  668.751µs, 950.712µs, 937.363µs, 1.071ms, 1.12ms, 1.248ms, 10.747ms
Bytes In      [total, mean]                     960000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)
