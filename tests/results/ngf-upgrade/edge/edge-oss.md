# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3a08fdafadfe0fb4a9c25679da1a1fcd6b181474
- Date: 2024-10-15T13:45:52Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1014001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 879.857µs
Latencies     [min, mean, 50, 90, 95, 99, max]  472.733µs, 846.68µs, 831.757µs, 948.108µs, 989.426µs, 1.13ms, 12.461ms
Bytes In      [total, mean]                     968005, 161.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-oss.png](http-oss.png)

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.99s, 802.597µs
Latencies     [min, mean, 50, 90, 95, 99, max]  646.501µs, 872.011µs, 851.784µs, 961.79µs, 1.006ms, 1.137ms, 12.519ms
Bytes In      [total, mean]                     930000, 155.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-oss.png](https-oss.png)
