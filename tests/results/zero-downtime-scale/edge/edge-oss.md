# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: fed4239ecb35f937b66bba7bd68d6894ca0762b3
- Date: 2024-11-01T00:13:12Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1355000
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.014ms
Latencies     [min, mean, 50, 90, 95, 99, max]  430.94µs, 901.297µs, 888.738µs, 1.037ms, 1.096ms, 1.467ms, 14.997ms
Bytes In      [total, mean]                     4775931, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.028ms
Latencies     [min, mean, 50, 90, 95, 99, max]  466.425µs, 941.352µs, 923.486µs, 1.085ms, 1.151ms, 1.54ms, 16.543ms
Bytes In      [total, mean]                     4595966, 153.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 1.159ms
Latencies     [min, mean, 50, 90, 95, 99, max]  400.634µs, 884.166µs, 883.021µs, 1.016ms, 1.068ms, 1.367ms, 14.634ms
Bytes In      [total, mean]                     7641643, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 1.167ms
Latencies     [min, mean, 50, 90, 95, 99, max]  427.567µs, 913.04µs, 906.062µs, 1.043ms, 1.094ms, 1.391ms, 14.792ms
Bytes In      [total, mean]                     7353699, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-oss.png](gradual-scale-down-affinity-https-oss.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 981.098µs
Latencies     [min, mean, 50, 90, 95, 99, max]  445.354µs, 874.943µs, 875.578µs, 1.011ms, 1.055ms, 1.262ms, 12.025ms
Bytes In      [total, mean]                     1910499, 159.21
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.019ms
Latencies     [min, mean, 50, 90, 95, 99, max]  434.583µs, 903.9µs, 895.883µs, 1.036ms, 1.087ms, 1.327ms, 8.382ms
Bytes In      [total, mean]                     1838386, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-oss.png](abrupt-scale-up-affinity-https-oss.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 998.28µs
Latencies     [min, mean, 50, 90, 95, 99, max]  481.507µs, 918.402µs, 909.478µs, 1.065ms, 1.125ms, 1.323ms, 13.656ms
Bytes In      [total, mean]                     1838395, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 975.08µs
Latencies     [min, mean, 50, 90, 95, 99, max]  441.227µs, 881.734µs, 879.226µs, 1.022ms, 1.082ms, 1.247ms, 7.957ms
Bytes In      [total, mean]                     1910435, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 908.983µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.479µs, 916.257µs, 903.315µs, 1.055ms, 1.117ms, 1.468ms, 26.462ms
Bytes In      [total, mean]                     4785334, 159.51
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-oss.png](gradual-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1ms
Latencies     [min, mean, 50, 90, 95, 99, max]  464.946µs, 943.129µs, 922.332µs, 1.082ms, 1.149ms, 1.502ms, 17.133ms
Bytes In      [total, mean]                     4616947, 153.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 814.931µs
Latencies     [min, mean, 50, 90, 95, 99, max]  417.756µs, 904.791µs, 896.338µs, 1.052ms, 1.115ms, 1.383ms, 18.578ms
Bytes In      [total, mean]                     15312004, 159.50
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-oss.png](gradual-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 826.007µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.525µs, 931.418µs, 917.297µs, 1.084ms, 1.152ms, 1.421ms, 19.343ms
Bytes In      [total, mean]                     14774252, 153.90
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.025ms
Latencies     [min, mean, 50, 90, 95, 99, max]  467.197µs, 926.232µs, 912.766µs, 1.084ms, 1.153ms, 1.385ms, 11.658ms
Bytes In      [total, mean]                     1846810, 153.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.024ms
Latencies     [min, mean, 50, 90, 95, 99, max]  459.536µs, 912.774µs, 904.507µs, 1.066ms, 1.129ms, 1.327ms, 10.956ms
Bytes In      [total, mean]                     1913953, 159.50
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 839.452µs
Latencies     [min, mean, 50, 90, 95, 99, max]  452.018µs, 882.796µs, 883.712µs, 1.013ms, 1.058ms, 1.2ms, 4.918ms
Bytes In      [total, mean]                     1913861, 159.49
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.12ms
Latencies     [min, mean, 50, 90, 95, 99, max]  473.095µs, 904.921µs, 902.088µs, 1.042ms, 1.093ms, 1.269ms, 7.449ms
Bytes In      [total, mean]                     1846961, 153.91
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)
