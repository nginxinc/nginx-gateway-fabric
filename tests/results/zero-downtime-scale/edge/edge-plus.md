# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 9a85dbcc0797e31557a3731688795aa166ee0f96
- Date: 2024-08-13T21:12:05Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1326000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 803.719µs
Latencies     [min, mean, 50, 90, 95, 99, max]  383.64µs, 813.964µs, 806.99µs, 953.502µs, 1.013ms, 1.27ms, 14.101ms
Bytes In      [total, mean]                     4826983, 160.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-plus.png](gradual-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 758.257µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.535µs, 830.965µs, 819.716µs, 956.264µs, 1.012ms, 1.271ms, 12.233ms
Bytes In      [total, mean]                     4619929, 154.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 792.628µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.338µs, 838.042µs, 828.579µs, 975.489µs, 1.034ms, 1.254ms, 34.701ms
Bytes In      [total, mean]                     7391971, 154.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 576.387µs
Latencies     [min, mean, 50, 90, 95, 99, max]  398.939µs, 808.743µs, 805.28µs, 944.663µs, 999.475µs, 1.211ms, 24.434ms
Bytes In      [total, mean]                     7723399, 160.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 959.085µs
Latencies     [min, mean, 50, 90, 95, 99, max]  407.604µs, 790.353µs, 787.313µs, 914.084µs, 960.312µs, 1.128ms, 5.357ms
Bytes In      [total, mean]                     1930730, 160.89
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 818.125µs
Latencies     [min, mean, 50, 90, 95, 99, max]  431.008µs, 848.722µs, 832.66µs, 989.584µs, 1.06ms, 1.302ms, 7.374ms
Bytes In      [total, mean]                     1848037, 154.00
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
Duration      [total, attack, wait]             2m0s, 2m0s, 863.199µs
Latencies     [min, mean, 50, 90, 95, 99, max]  418.492µs, 803.518µs, 806.395µs, 939.15µs, 987.692µs, 1.131ms, 4.68ms
Bytes In      [total, mean]                     1930853, 160.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 551.15µs
Latencies     [min, mean, 50, 90, 95, 99, max]  440.591µs, 827.614µs, 825.911µs, 956.666µs, 1.003ms, 1.144ms, 5.246ms
Bytes In      [total, mean]                     1848038, 154.00
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
Duration      [total, attack, wait]             5m0s, 5m0s, 1.085ms
Latencies     [min, mean, 50, 90, 95, 99, max]  432.999µs, 868.073µs, 848.532µs, 1.006ms, 1.082ms, 1.405ms, 11.871ms
Bytes In      [total, mean]                     4623138, 154.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 718.592µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.631µs, 835.92µs, 824.014µs, 974.867µs, 1.045ms, 1.341ms, 11.582ms
Bytes In      [total, mean]                     4832889, 161.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 884.796µs
Latencies     [min, mean, 50, 90, 95, 99, max]  393.131µs, 822.987µs, 817.679µs, 949.34µs, 1.002ms, 1.257ms, 43.09ms
Bytes In      [total, mean]                     15465360, 161.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 783.867µs
Latencies     [min, mean, 50, 90, 95, 99, max]  414.755µs, 853.381µs, 841.866µs, 984.689µs, 1.041ms, 1.291ms, 33.395ms
Bytes In      [total, mean]                     14793566, 154.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 805.973µs
Latencies     [min, mean, 50, 90, 95, 99, max]  384.007µs, 820.512µs, 818.95µs, 953.165µs, 1.009ms, 1.223ms, 8.993ms
Bytes In      [total, mean]                     1933211, 161.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 833.982µs
Latencies     [min, mean, 50, 90, 95, 99, max]  446.32µs, 839.644µs, 833.205µs, 961.082µs, 1.009ms, 1.202ms, 12.421ms
Bytes In      [total, mean]                     1849167, 154.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 860.874µs
Latencies     [min, mean, 50, 90, 95, 99, max]  417.841µs, 831.148µs, 831.149µs, 959.406µs, 1.005ms, 1.141ms, 4.048ms
Bytes In      [total, mean]                     1933278, 161.11
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 834.483µs
Latencies     [min, mean, 50, 90, 95, 99, max]  446.663µs, 846.991µs, 845.434µs, 977.288µs, 1.022ms, 1.168ms, 5.424ms
Bytes In      [total, mean]                     1849173, 154.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

