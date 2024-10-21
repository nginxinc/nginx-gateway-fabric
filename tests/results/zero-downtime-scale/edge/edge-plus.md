# Results

## Test environment

NGINX Plus: true

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

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 882.701µs
Latencies     [min, mean, 50, 90, 95, 99, max]  447.018µs, 887.856µs, 880.209µs, 1.014ms, 1.066ms, 1.34ms, 12.115ms
Bytes In      [total, mean]                     4655980, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 883.526µs
Latencies     [min, mean, 50, 90, 95, 99, max]  414.12µs, 858.377µs, 857.716µs, 987.718µs, 1.036ms, 1.323ms, 12.309ms
Bytes In      [total, mean]                     4835948, 161.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 645.314µs
Latencies     [min, mean, 50, 90, 95, 99, max]  404.11µs, 855.977µs, 856.051µs, 984.687µs, 1.034ms, 1.254ms, 23.487ms
Bytes In      [total, mean]                     7737643, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 882.826µs
Latencies     [min, mean, 50, 90, 95, 99, max]  428.641µs, 876.809µs, 872.998µs, 1.007ms, 1.059ms, 1.28ms, 12.573ms
Bytes In      [total, mean]                     7449735, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 844.781µs
Latencies     [min, mean, 50, 90, 95, 99, max]  432.221µs, 855.933µs, 855.366µs, 979.815µs, 1.025ms, 1.228ms, 10.512ms
Bytes In      [total, mean]                     1934334, 161.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.046ms
Latencies     [min, mean, 50, 90, 95, 99, max]  440.703µs, 880.516µs, 881.193µs, 1.007ms, 1.057ms, 1.253ms, 6.875ms
Bytes In      [total, mean]                     1862351, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 961.936µs
Latencies     [min, mean, 50, 90, 95, 99, max]  405.919µs, 861.927µs, 865.96µs, 992.282µs, 1.039ms, 1.167ms, 36.375ms
Bytes In      [total, mean]                     1934386, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 873.391µs
Latencies     [min, mean, 50, 90, 95, 99, max]  447.373µs, 888.896µs, 885.533µs, 1.015ms, 1.062ms, 1.217ms, 36.382ms
Bytes In      [total, mean]                     1862425, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 888.736µs
Latencies     [min, mean, 50, 90, 95, 99, max]  459.383µs, 903.921µs, 890.521µs, 1.035ms, 1.091ms, 1.412ms, 13.05ms
Bytes In      [total, mean]                     4656092, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 968.944µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.765µs, 877.932µs, 872.033µs, 1.01ms, 1.069ms, 1.404ms, 11.75ms
Bytes In      [total, mean]                     4835992, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 868.25µs
Latencies     [min, mean, 50, 90, 95, 99, max]  436.337µs, 904.839µs, 898.798µs, 1.036ms, 1.091ms, 1.352ms, 11.98ms
Bytes In      [total, mean]                     14899102, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 877.384µs
Latencies     [min, mean, 50, 90, 95, 99, max]  429.428µs, 877.117µs, 875.557µs, 1.008ms, 1.06ms, 1.325ms, 11.959ms
Bytes In      [total, mean]                     15475290, 161.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 847.027µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.405µs, 881.467µs, 882.133µs, 1.011ms, 1.058ms, 1.296ms, 12.033ms
Bytes In      [total, mean]                     1934415, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 844.179µs
Latencies     [min, mean, 50, 90, 95, 99, max]  479.416µs, 916.265µs, 910.635µs, 1.046ms, 1.098ms, 1.346ms, 7.79ms
Bytes In      [total, mean]                     1862450, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 771.524µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.735µs, 884.556µs, 890.309µs, 1.014ms, 1.057ms, 1.197ms, 4.981ms
Bytes In      [total, mean]                     1934345, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 862.724µs
Latencies     [min, mean, 50, 90, 95, 99, max]  497.824µs, 914.248µs, 913.952µs, 1.042ms, 1.09ms, 1.237ms, 12.169ms
Bytes In      [total, mean]                     1862409, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)
