# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 830.225µs
Latencies     [min, mean, 50, 90, 95, 99, max]  643.941µs, 905.713µs, 889.406µs, 1.029ms, 1.074ms, 1.218ms, 12.298ms
Bytes In      [total, mean]                     912000, 152.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https-plus.png](https-plus.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.991s, 59.991s, 727.47µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.538µs, 824.665µs, 805.849µs, 940.169µs, 985.195µs, 1.133ms, 22.968ms
Bytes In      [total, mean]                     955990, 159.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http-plus.png](http-plus.png)
