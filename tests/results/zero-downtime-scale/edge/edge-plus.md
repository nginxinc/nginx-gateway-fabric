# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: bf8ea47203eb4695af0d359243c73de2d1badbbf
- Date: 2024-09-13T20:33:11Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1639000
- vCPUs per node: 16
- RAM per node: 65853960Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 838.654µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.793µs, 895.22µs, 888.223µs, 1.02ms, 1.072ms, 1.357ms, 23.439ms
Bytes In      [total, mean]                     4677038, 155.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 817.673µs
Latencies     [min, mean, 50, 90, 95, 99, max]  443.502µs, 875.28µs, 869.666µs, 996.17µs, 1.048ms, 1.34ms, 23.496ms
Bytes In      [total, mean]                     4854041, 161.80
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-plus.png](gradual-scale-up-affinity-http-plus.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 934.141µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.386µs, 862.218µs, 860.273µs, 993.684µs, 1.045ms, 1.305ms, 10.817ms
Bytes In      [total, mean]                     7766579, 161.80
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 936.269µs
Latencies     [min, mean, 50, 90, 95, 99, max]  420.305µs, 885.913µs, 879.821µs, 1.017ms, 1.074ms, 1.363ms, 12.081ms
Bytes In      [total, mean]                     7483128, 155.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 825.237µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.923µs, 881.321µs, 874.85µs, 1.017ms, 1.07ms, 1.269ms, 12.888ms
Bytes In      [total, mean]                     1870817, 155.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 811.898µs
Latencies     [min, mean, 50, 90, 95, 99, max]  433.705µs, 859.014µs, 855.668µs, 993.424µs, 1.042ms, 1.222ms, 11.175ms
Bytes In      [total, mean]                     1941620, 161.80
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 951.106µs
Latencies     [min, mean, 50, 90, 95, 99, max]  465.627µs, 892.868µs, 882.75µs, 1.027ms, 1.079ms, 1.252ms, 26.823ms
Bytes In      [total, mean]                     1870707, 155.89
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 822.955µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.669µs, 846.849µs, 846.377µs, 987.528µs, 1.037ms, 1.201ms, 26.852ms
Bytes In      [total, mean]                     1941643, 161.80
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.005ms
Latencies     [min, mean, 50, 90, 95, 99, max]  432.267µs, 869.923µs, 862.861µs, 994.904µs, 1.046ms, 1.395ms, 8.753ms
Bytes In      [total, mean]                     4860067, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 897.953µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.595µs, 892.323µs, 883.737µs, 1.018ms, 1.074ms, 1.381ms, 10.233ms
Bytes In      [total, mean]                     4680027, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 858.592µs
Latencies     [min, mean, 50, 90, 95, 99, max]  441.716µs, 879.772µs, 872.168µs, 1.004ms, 1.056ms, 1.328ms, 13.475ms
Bytes In      [total, mean]                     14976225, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 648.016µs
Latencies     [min, mean, 50, 90, 95, 99, max]  411.621µs, 856.204µs, 854.93µs, 983.717µs, 1.032ms, 1.288ms, 13.942ms
Bytes In      [total, mean]                     15552181, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 827.87µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.721µs, 872.176µs, 871.932µs, 999.479µs, 1.043ms, 1.27ms, 5.779ms
Bytes In      [total, mean]                     1944021, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 852.421µs
Latencies     [min, mean, 50, 90, 95, 99, max]  483.676µs, 893.114µs, 888.653µs, 1.021ms, 1.069ms, 1.317ms, 6.032ms
Bytes In      [total, mean]                     1872026, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 974.014µs
Latencies     [min, mean, 50, 90, 95, 99, max]  502.218µs, 910.536µs, 904.281µs, 1.032ms, 1.083ms, 1.248ms, 32.343ms
Bytes In      [total, mean]                     1872022, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 752.857µs
Latencies     [min, mean, 50, 90, 95, 99, max]  464.305µs, 882.383µs, 882.881µs, 1.013ms, 1.061ms, 1.225ms, 8.445ms
Bytes In      [total, mean]                     1943988, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)
