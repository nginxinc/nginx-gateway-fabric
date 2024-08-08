# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 809c0838e2f2658c3c4cd48325ffb0bc5a92a002
- Date: 2024-08-08T18:03:35Z
- Dirty: false

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
Duration      [total, attack, wait]             5m0s, 5m0s, 920.383µs
Latencies     [min, mean, 50, 90, 95, 99, max]  467.39µs, 915.759µs, 904.095µs, 1.038ms, 1.091ms, 1.399ms, 12.455ms
Bytes In      [total, mean]                     4566107, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 888.362µs
Latencies     [min, mean, 50, 90, 95, 99, max]  424.554µs, 892.62µs, 885.293µs, 1.017ms, 1.066ms, 1.389ms, 12.406ms
Bytes In      [total, mean]                     4776036, 159.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 742.1µs
Latencies     [min, mean, 50, 90, 95, 99, max]  402.781µs, 854.355µs, 854.015µs, 983.84µs, 1.031ms, 1.266ms, 12.36ms
Bytes In      [total, mean]                     7641568, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 803.047µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.601µs, 883.996µs, 877.173µs, 1.013ms, 1.067ms, 1.31ms, 12.143ms
Bytes In      [total, mean]                     7305561, 152.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 721.022µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.463µs, 841.335µs, 837.587µs, 963.799µs, 1.012ms, 1.261ms, 11.902ms
Bytes In      [total, mean]                     1910340, 159.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 827.402µs
Latencies     [min, mean, 50, 90, 95, 99, max]  446.551µs, 849.947µs, 850.66µs, 975.351µs, 1.018ms, 1.157ms, 5.312ms
Bytes In      [total, mean]                     1826382, 152.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 769.605µs
Latencies     [min, mean, 50, 90, 95, 99, max]  385.773µs, 822.503µs, 826.553µs, 951.26µs, 995.15µs, 1.13ms, 18.169ms
Bytes In      [total, mean]                     1910446, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 820.762µs
Latencies     [min, mean, 50, 90, 95, 99, max]  430.981µs, 857.686µs, 856.215µs, 980.989µs, 1.028ms, 1.163ms, 19.103ms
Bytes In      [total, mean]                     1826468, 152.21
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
Duration      [total, attack, wait]             5m0s, 5m0s, 839.173µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.246µs, 916.945µs, 905.949µs, 1.045ms, 1.103ms, 1.395ms, 11.457ms
Bytes In      [total, mean]                     4574963, 152.50
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 851.566µs
Latencies     [min, mean, 50, 90, 95, 99, max]  433.085µs, 891.409µs, 887.337µs, 1.02ms, 1.071ms, 1.383ms, 23.41ms
Bytes In      [total, mean]                     4775841, 159.19
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
Duration      [total, attack, wait]             16m0s, 16m0s, 817.648µs
Latencies     [min, mean, 50, 90, 95, 99, max]  436.056µs, 898.531µs, 882.067µs, 1.017ms, 1.072ms, 1.337ms, 1.008s
Bytes In      [total, mean]                     14640055, 152.50
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 795.046µs
Latencies     [min, mean, 50, 90, 95, 99, max]  388.919µs, 859.284µs, 858.392µs, 987.158µs, 1.035ms, 1.289ms, 20.787ms
Bytes In      [total, mean]                     15283391, 159.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 767.523µs
Latencies     [min, mean, 50, 90, 95, 99, max]  464.09µs, 881.037µs, 876.736µs, 1.002ms, 1.049ms, 1.278ms, 10.93ms
Bytes In      [total, mean]                     1830002, 152.50
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 934.271µs
Latencies     [min, mean, 50, 90, 95, 99, max]  417.478µs, 855.503µs, 857.623µs, 978.604µs, 1.026ms, 1.229ms, 4.693ms
Bytes In      [total, mean]                     1910431, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 826.982µs
Latencies     [min, mean, 50, 90, 95, 99, max]  445.393µs, 867.121µs, 875.212µs, 995.198µs, 1.036ms, 1.157ms, 8.939ms
Bytes In      [total, mean]                     1910413, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 812.474µs
Latencies     [min, mean, 50, 90, 95, 99, max]  480.681µs, 904.284µs, 903.609µs, 1.031ms, 1.078ms, 1.231ms, 8.084ms
Bytes In      [total, mean]                     1830010, 152.50
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

