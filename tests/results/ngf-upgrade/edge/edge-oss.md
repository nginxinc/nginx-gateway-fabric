# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855004Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 977.211µs
Latencies     [min, mean, 50, 90, 95, 99, max]  632.106µs, 874.17µs, 856.589µs, 998.899µs, 1.049ms, 1.192ms, 7.44ms
Bytes In      [total, mean]                     966000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 843.606µs
Latencies     [min, mean, 50, 90, 95, 99, max]  615.636µs, 858.807µs, 842.243µs, 963.533µs, 1.014ms, 1.166ms, 7.665ms
Bytes In      [total, mean]                     927944, 154.66
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)
