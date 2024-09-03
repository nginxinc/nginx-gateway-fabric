# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855004Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 3.965ms
Latencies     [min, mean, 50, 90, 95, 99, max]  404.177µs, 877.278µs, 870.227µs, 1.009ms, 1.064ms, 1.339ms, 12.386ms
Bytes In      [total, mean]                     4626000, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 2.248ms
Latencies     [min, mean, 50, 90, 95, 99, max]  381.381µs, 846.858µs, 843.817µs, 980.732µs, 1.032ms, 1.269ms, 12.219ms
Bytes In      [total, mean]                     4835967, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 694.581µs
Latencies     [min, mean, 50, 90, 95, 99, max]  380.103µs, 838.63µs, 838.673µs, 966.853µs, 1.013ms, 1.245ms, 18.359ms
Bytes In      [total, mean]                     7737784, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 1.051ms
Latencies     [min, mean, 50, 90, 95, 99, max]  432.878µs, 861.821µs, 856.941µs, 985.9µs, 1.034ms, 1.237ms, 34.981ms
Bytes In      [total, mean]                     7401698, 154.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.093ms
Latencies     [min, mean, 50, 90, 95, 99, max]  430.427µs, 886.722µs, 878.744µs, 1.038ms, 1.096ms, 1.305ms, 5.263ms
Bytes In      [total, mean]                     1850350, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-oss.png](abrupt-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 932.099µs
Latencies     [min, mean, 50, 90, 95, 99, max]  412.823µs, 840.224µs, 842.706µs, 982.988µs, 1.035ms, 1.239ms, 4.859ms
Bytes In      [total, mean]                     1934382, 161.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 939.323µs
Latencies     [min, mean, 50, 90, 95, 99, max]  476.512µs, 912.927µs, 913.927µs, 1.043ms, 1.089ms, 1.24ms, 5.939ms
Bytes In      [total, mean]                     1850421, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 917.61µs
Latencies     [min, mean, 50, 90, 95, 99, max]  432.34µs, 883.12µs, 886.593µs, 1.018ms, 1.065ms, 1.224ms, 4.956ms
Bytes In      [total, mean]                     1934463, 161.21
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
Duration      [total, attack, wait]             5m0s, 5m0s, 853.431µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.011µs, 885.307µs, 871.566µs, 1.02ms, 1.08ms, 1.377ms, 12.314ms
Bytes In      [total, mean]                     4626089, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.118ms
Latencies     [min, mean, 50, 90, 95, 99, max]  429.304µs, 853.539µs, 845.965µs, 987.293µs, 1.042ms, 1.34ms, 7.511ms
Bytes In      [total, mean]                     4835978, 161.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 886.903µs
Latencies     [min, mean, 50, 90, 95, 99, max]  397.498µs, 845.962µs, 839.184µs, 985.753µs, 1.044ms, 1.318ms, 38.887ms
Bytes In      [total, mean]                     15475093, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-oss.png](gradual-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.008ms
Latencies     [min, mean, 50, 90, 95, 99, max]  422.557µs, 871.52µs, 862.16µs, 1.007ms, 1.066ms, 1.326ms, 28.054ms
Bytes In      [total, mean]                     14803380, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 792.151µs
Latencies     [min, mean, 50, 90, 95, 99, max]  411.356µs, 865.528µs, 858.649µs, 1.008ms, 1.065ms, 1.257ms, 9.083ms
Bytes In      [total, mean]                     1850394, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 792.949µs
Latencies     [min, mean, 50, 90, 95, 99, max]  445.048µs, 848.311µs, 843.896µs, 983.226µs, 1.031ms, 1.204ms, 7.715ms
Bytes In      [total, mean]                     1934355, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 928.94µs
Latencies     [min, mean, 50, 90, 95, 99, max]  456.968µs, 910.159µs, 903.343µs, 1.068ms, 1.122ms, 1.289ms, 5.406ms
Bytes In      [total, mean]                     1850411, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 905.042µs
Latencies     [min, mean, 50, 90, 95, 99, max]  409.244µs, 872.655µs, 871.308µs, 1.025ms, 1.077ms, 1.239ms, 3.424ms
Bytes In      [total, mean]                     1934426, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)
