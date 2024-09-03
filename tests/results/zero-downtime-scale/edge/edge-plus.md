# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 81eef156ceeefa3997d25d023a772d1201109583
- Date: 2024-09-03T20:06:12Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
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
Duration      [total, attack, wait]             5m0s, 5m0s, 845.308µs
Latencies     [min, mean, 50, 90, 95, 99, max]  408.117µs, 863.236µs, 847.209µs, 993.524µs, 1.054ms, 1.407ms, 12.856ms
Bytes In      [total, mean]                     4596078, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 837.496µs
Latencies     [min, mean, 50, 90, 95, 99, max]  414.66µs, 839.84µs, 826.498µs, 968.692µs, 1.031ms, 1.399ms, 12.725ms
Bytes In      [total, mean]                     4805920, 160.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 767.235µs
Latencies     [min, mean, 50, 90, 95, 99, max]  399.638µs, 813.04µs, 811.182µs, 936.089µs, 983.41µs, 1.259ms, 12.42ms
Bytes In      [total, mean]                     7689553, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 524.916µs
Latencies     [min, mean, 50, 90, 95, 99, max]  413.88µs, 842.409µs, 836.074µs, 966.876µs, 1.02ms, 1.292ms, 12.081ms
Bytes In      [total, mean]                     7353529, 153.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 772.297µs
Latencies     [min, mean, 50, 90, 95, 99, max]  416.985µs, 847µs, 852.87µs, 972.724µs, 1.015ms, 1.2ms, 6.576ms
Bytes In      [total, mean]                     1922406, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 813.233µs
Latencies     [min, mean, 50, 90, 95, 99, max]  445.026µs, 869.778µs, 860.948µs, 994.839µs, 1.057ms, 1.322ms, 12.293ms
Bytes In      [total, mean]                     1838411, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 877.993µs
Latencies     [min, mean, 50, 90, 95, 99, max]  465.189µs, 872.49µs, 869.604µs, 997.771µs, 1.042ms, 1.217ms, 29.13ms
Bytes In      [total, mean]                     1838442, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 849.903µs
Latencies     [min, mean, 50, 90, 95, 99, max]  414.498µs, 846.53µs, 849.31µs, 969.394µs, 1.007ms, 1.153ms, 34.305ms
Bytes In      [total, mean]                     1922434, 160.20
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
Duration      [total, attack, wait]             5m0s, 5m0s, 587.754µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.093µs, 885.511µs, 872.04µs, 1.01ms, 1.07ms, 1.369ms, 12.266ms
Bytes In      [total, mean]                     4601987, 153.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 720.008µs
Latencies     [min, mean, 50, 90, 95, 99, max]  424.068µs, 862.754µs, 855.401µs, 986.983µs, 1.044ms, 1.362ms, 11.966ms
Bytes In      [total, mean]                     4806020, 160.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 975.101µs
Latencies     [min, mean, 50, 90, 95, 99, max]  381.609µs, 861µs, 854.707µs, 1.001ms, 1.068ms, 1.399ms, 31.856ms
Bytes In      [total, mean]                     15379089, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.01ms
Latencies     [min, mean, 50, 90, 95, 99, max]  427.924µs, 888.055µs, 875.424µs, 1.027ms, 1.098ms, 1.431ms, 47.717ms
Bytes In      [total, mean]                     14726285, 153.40
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
Duration      [total, attack, wait]             2m0s, 2m0s, 794.831µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.008µs, 845.661µs, 842.635µs, 984.235µs, 1.045ms, 1.294ms, 12.298ms
Bytes In      [total, mean]                     1922418, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 929.444µs
Latencies     [min, mean, 50, 90, 95, 99, max]  432.284µs, 880.427µs, 871.277µs, 1.032ms, 1.097ms, 1.4ms, 8.092ms
Bytes In      [total, mean]                     1840798, 153.40
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.323ms
Latencies     [min, mean, 50, 90, 95, 99, max]  451.618µs, 902.897µs, 893.019µs, 1.058ms, 1.116ms, 1.295ms, 11.002ms
Bytes In      [total, mean]                     1840821, 153.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 595.632µs
Latencies     [min, mean, 50, 90, 95, 99, max]  473.709µs, 880.573µs, 873.747µs, 1.021ms, 1.074ms, 1.303ms, 10.946ms
Bytes In      [total, mean]                     1922421, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)
