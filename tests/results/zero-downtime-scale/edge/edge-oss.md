# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3c029b1417c1f89f2a29aeef07f47078640e28b2
- Date: 2024-08-15T00:04:25Z
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
Duration      [total, attack, wait]             5m0s, 5m0s, 747.767µs
Latencies     [min, mean, 50, 90, 95, 99, max]  414.569µs, 871.361µs, 860.543µs, 999.408µs, 1.053ms, 1.372ms, 23.601ms
Bytes In      [total, mean]                     4806003, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 909.264µs
Latencies     [min, mean, 50, 90, 95, 99, max]  433.814µs, 887.623µs, 877.584µs, 1.018ms, 1.071ms, 1.351ms, 31.021ms
Bytes In      [total, mean]                     4596080, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 781.11µs
Latencies     [min, mean, 50, 90, 95, 99, max]  398.298µs, 843.922µs, 844.976µs, 978.632µs, 1.024ms, 1.248ms, 11.206ms
Bytes In      [total, mean]                     7689685, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 807.808µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.654µs, 873.586µs, 868.462µs, 1.006ms, 1.057ms, 1.275ms, 12.377ms
Bytes In      [total, mean]                     7353559, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-oss.png](gradual-scale-down-affinity-https-oss.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 928.789µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.82µs, 898.446µs, 888.93µs, 1.04ms, 1.101ms, 1.277ms, 10.922ms
Bytes In      [total, mean]                     1838365, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-oss.png](abrupt-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 946.422µs
Latencies     [min, mean, 50, 90, 95, 99, max]  446.889µs, 863.531µs, 864.643µs, 993.534µs, 1.037ms, 1.24ms, 2.894ms
Bytes In      [total, mean]                     1922331, 160.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 803.302µs
Latencies     [min, mean, 50, 90, 95, 99, max]  463.744µs, 843.899µs, 846.272µs, 971.262µs, 1.014ms, 1.138ms, 6.961ms
Bytes In      [total, mean]                     1922383, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 519.961µs
Latencies     [min, mean, 50, 90, 95, 99, max]  451.477µs, 865.661µs, 863.067µs, 996.552µs, 1.046ms, 1.178ms, 6.933ms
Bytes In      [total, mean]                     1838348, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.037ms
Latencies     [min, mean, 50, 90, 95, 99, max]  462.784µs, 893.425µs, 884.052µs, 1.015ms, 1.068ms, 1.356ms, 11.968ms
Bytes In      [total, mean]                     4595998, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 954.461µs
Latencies     [min, mean, 50, 90, 95, 99, max]  434.113µs, 868.429µs, 864.548µs, 996.703µs, 1.052ms, 1.346ms, 9.717ms
Bytes In      [total, mean]                     4805945, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-oss.png](gradual-scale-up-http-oss.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 871.349µs
Latencies     [min, mean, 50, 90, 95, 99, max]  440.552µs, 888.19µs, 880.655µs, 1.018ms, 1.073ms, 1.313ms, 41.155ms
Bytes In      [total, mean]                     14707143, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 871.306µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.257µs, 860.433µs, 858.936µs, 992.074µs, 1.043ms, 1.276ms, 29.859ms
Bytes In      [total, mean]                     15379205, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-oss.png](gradual-scale-down-http-oss.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 995.727µs
Latencies     [min, mean, 50, 90, 95, 99, max]  462.116µs, 902.223µs, 898.476µs, 1.03ms, 1.082ms, 1.307ms, 8.366ms
Bytes In      [total, mean]                     1922414, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 979.865µs
Latencies     [min, mean, 50, 90, 95, 99, max]  483.569µs, 921.536µs, 912.988µs, 1.047ms, 1.097ms, 1.324ms, 11.19ms
Bytes In      [total, mean]                     1838405, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 909.807µs
Latencies     [min, mean, 50, 90, 95, 99, max]  471.254µs, 901.676µs, 904.018µs, 1.034ms, 1.079ms, 1.235ms, 3.541ms
Bytes In      [total, mean]                     1922315, 160.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.074ms
Latencies     [min, mean, 50, 90, 95, 99, max]  523.474µs, 933.653µs, 926.079µs, 1.068ms, 1.121ms, 1.276ms, 10.334ms
Bytes In      [total, mean]                     1838363, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)
