# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 12
- k8s version: v1.28.8-gke.1095000
- vCPUs per node: 16
- RAM per node: 65855088Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 1.192ms
Latencies     [min, mean, 50, 90, 95, 99, max]  867.75µs, 1.215ms, 1.14ms, 1.295ms, 1.357ms, 1.965ms, 24.186ms
Bytes In      [total, mean]                     914021, 152.34
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![https.png](https.png)

## Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         6000, 100.02, 100.01
Duration      [total, attack, wait]             59.992s, 59.991s, 996.367µs
Latencies     [min, mean, 50, 90, 95, 99, max]  824.305µs, 1.172ms, 1.122ms, 1.269ms, 1.329ms, 1.685ms, 13.407ms
Bytes In      [total, mean]                     954000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  
Error Set:
```

![http.png](http.png)
