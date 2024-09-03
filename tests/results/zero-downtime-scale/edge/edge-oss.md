# Results

## Test environment

NGINX Plus: false

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

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 912.577µs
Latencies     [min, mean, 50, 90, 95, 99, max]  449.563µs, 895.986µs, 882.939µs, 1.013ms, 1.063ms, 1.399ms, 14.696ms
Bytes In      [total, mean]                     4776076, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 972.143µs
Latencies     [min, mean, 50, 90, 95, 99, max]  452.594µs, 922.96µs, 902.912µs, 1.04ms, 1.095ms, 1.436ms, 14.015ms
Bytes In      [total, mean]                     4566030, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 827.44µs
Latencies     [min, mean, 50, 90, 95, 99, max]  457.186µs, 892.168µs, 883.537µs, 1.021ms, 1.075ms, 1.313ms, 19.185ms
Bytes In      [total, mean]                     7305597, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-oss.png](gradual-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 824.449µs
Latencies     [min, mean, 50, 90, 95, 99, max]  416.541µs, 861.311µs, 858.551µs, 991.818µs, 1.041ms, 1.285ms, 16.588ms
Bytes In      [total, mean]                     7641646, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 887.798µs
Latencies     [min, mean, 50, 90, 95, 99, max]  461.665µs, 894.274µs, 886.585µs, 1.03ms, 1.082ms, 1.284ms, 9.876ms
Bytes In      [total, mean]                     1826332, 152.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-oss.png](abrupt-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1ms
Latencies     [min, mean, 50, 90, 95, 99, max]  424.161µs, 858.558µs, 855.766µs, 983.596µs, 1.032ms, 1.24ms, 11.042ms
Bytes In      [total, mean]                     1910373, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 854.961µs
Latencies     [min, mean, 50, 90, 95, 99, max]  461.232µs, 906.167µs, 901.375µs, 1.047ms, 1.097ms, 1.251ms, 12.927ms
Bytes In      [total, mean]                     1826411, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 821.843µs
Latencies     [min, mean, 50, 90, 95, 99, max]  449.521µs, 885.12µs, 884.147µs, 1.023ms, 1.071ms, 1.215ms, 7.313ms
Bytes In      [total, mean]                     1910395, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.205ms
Latencies     [min, mean, 50, 90, 95, 99, max]  448.442µs, 904.375µs, 893.672µs, 1.031ms, 1.082ms, 1.394ms, 12.504ms
Bytes In      [total, mean]                     4572069, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.005ms
Latencies     [min, mean, 50, 90, 95, 99, max]  421.273µs, 873.005µs, 869.126µs, 1.007ms, 1.062ms, 1.386ms, 12.47ms
Bytes In      [total, mean]                     4776016, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-oss.png](gradual-scale-up-http-oss.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 851.954µs
Latencies     [min, mean, 50, 90, 95, 99, max]  408.134µs, 890.732µs, 884.83µs, 1.028ms, 1.081ms, 1.315ms, 57.444ms
Bytes In      [total, mean]                     15283293, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-oss.png](gradual-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 920.048µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.753µs, 911.516µs, 900.3µs, 1.046ms, 1.103ms, 1.349ms, 58.615ms
Bytes In      [total, mean]                     14630033, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 813.779µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.953µs, 897.814µs, 893.81µs, 1.033ms, 1.086ms, 1.266ms, 3.297ms
Bytes In      [total, mean]                     1910398, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 797.823µs
Latencies     [min, mean, 50, 90, 95, 99, max]  451.905µs, 916.394µs, 908.107µs, 1.051ms, 1.102ms, 1.295ms, 11.815ms
Bytes In      [total, mean]                     1828832, 152.40
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
Duration      [total, attack, wait]             2m0s, 2m0s, 948.213µs
Latencies     [min, mean, 50, 90, 95, 99, max]  491.649µs, 905.043µs, 896.903µs, 1.031ms, 1.08ms, 1.213ms, 4.276ms
Bytes In      [total, mean]                     1910380, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.187ms
Latencies     [min, mean, 50, 90, 95, 99, max]  502.373µs, 935.742µs, 920.259µs, 1.06ms, 1.113ms, 1.235ms, 7.806ms
Bytes In      [total, mean]                     1828867, 152.41
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)
