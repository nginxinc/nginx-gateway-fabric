# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
- vCPUs per node: 16
- RAM per node: 65853964Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 868µs
Latencies     [min, mean, 50, 90, 95, 99, max]  673.831µs, 905.163µs, 871.482µs, 1.007ms, 1.058ms, 1.21ms, 12.461ms
Bytes In      [total, mean]                     926034, 154.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 738.617µs
Latencies     [min, mean, 50, 90, 95, 99, max]  611.455µs, 861.834µs, 832.286µs, 967.223µs, 1.017ms, 1.204ms, 12.717ms
Bytes In      [total, mean]                     960000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)
