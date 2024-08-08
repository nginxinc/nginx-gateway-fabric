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

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 886.845µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.76µs, 916.19µs, 889.909µs, 1.074ms, 1.15ms, 1.424ms, 16.721ms
Bytes In      [total, mean]                     4565944, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 806.14µs
Latencies     [min, mean, 50, 90, 95, 99, max]  413.867µs, 888.634µs, 864.596µs, 1.034ms, 1.108ms, 1.417ms, 13.323ms
Bytes In      [total, mean]                     4775951, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-plus.png](gradual-scale-up-affinity-http-plus.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 910.095µs
Latencies     [min, mean, 50, 90, 95, 99, max]  432.763µs, 896.672µs, 888.9µs, 1.045ms, 1.106ms, 1.336ms, 12.327ms
Bytes In      [total, mean]                     7305440, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 541.68µs
Latencies     [min, mean, 50, 90, 95, 99, max]  419.938µs, 865.374µs, 864.258µs, 1.008ms, 1.063ms, 1.265ms, 12.367ms
Bytes In      [total, mean]                     7641513, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 883.852µs
Latencies     [min, mean, 50, 90, 95, 99, max]  475.602µs, 906.791µs, 892.732µs, 1.049ms, 1.106ms, 1.304ms, 19.813ms
Bytes In      [total, mean]                     1826426, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 730.725µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.802µs, 864.033µs, 863.118µs, 1.006ms, 1.061ms, 1.254ms, 19.483ms
Bytes In      [total, mean]                     1910465, 159.21
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
Duration      [total, attack, wait]             2m0s, 2m0s, 862.745µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.516µs, 882.838µs, 880.811µs, 1.02ms, 1.067ms, 1.201ms, 11.995ms
Bytes In      [total, mean]                     1826453, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 953.093µs
Latencies     [min, mean, 50, 90, 95, 99, max]  401.333µs, 868.903µs, 863.347µs, 1.01ms, 1.072ms, 1.421ms, 11.988ms
Bytes In      [total, mean]                     1910397, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 924.541µs
Latencies     [min, mean, 50, 90, 95, 99, max]  478.692µs, 926.928µs, 914.83µs, 1.066ms, 1.126ms, 1.413ms, 10.691ms
Bytes In      [total, mean]                     4571975, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 687.968µs
Latencies     [min, mean, 50, 90, 95, 99, max]  424.623µs, 896.05µs, 888.997µs, 1.031ms, 1.09ms, 1.456ms, 12.284ms
Bytes In      [total, mean]                     4778966, 159.30
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
Duration      [total, attack, wait]             16m0s, 16m0s, 868.01µs
Latencies     [min, mean, 50, 90, 95, 99, max]  425.429µs, 926.439µs, 914.318µs, 1.077ms, 1.145ms, 1.413ms, 9.922ms
Bytes In      [total, mean]                     14630290, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.018ms
Latencies     [min, mean, 50, 90, 95, 99, max]  401.906µs, 895.592µs, 889.701µs, 1.043ms, 1.107ms, 1.391ms, 28.271ms
Bytes In      [total, mean]                     15292864, 159.30
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.158ms
Latencies     [min, mean, 50, 90, 95, 99, max]  484.491µs, 945.599µs, 928.892µs, 1.108ms, 1.176ms, 1.442ms, 15.819ms
Bytes In      [total, mean]                     1828861, 152.41
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 835.306µs
Latencies     [min, mean, 50, 90, 95, 99, max]  459.184µs, 924.014µs, 907.46µs, 1.081ms, 1.149ms, 1.364ms, 17.217ms
Bytes In      [total, mean]                     1911528, 159.29
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.00
Duration      [total, attack, wait]             2m0s, 2m0s, 7.503ms
Latencies     [min, mean, 50, 90, 95, 99, max]  521.027µs, 988.926µs, 973.869µs, 1.167ms, 1.23ms, 1.397ms, 12.339ms
Bytes In      [total, mean]                     1828769, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.253ms
Latencies     [min, mean, 50, 90, 95, 99, max]  471.165µs, 948.047µs, 942.446µs, 1.117ms, 1.176ms, 1.341ms, 6.867ms
Bytes In      [total, mean]                     1911640, 159.30
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

