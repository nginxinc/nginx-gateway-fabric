# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: bf8ea47203eb4695af0d359243c73de2d1badbbf
- Date: 2024-09-13T20:33:11Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1639000
- vCPUs per node: 16
- RAM per node: 65853968Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 941.282µs
Latencies     [min, mean, 50, 90, 95, 99, max]  415.473µs, 856.715µs, 838.478µs, 984.981µs, 1.047ms, 1.391ms, 12.998ms
Bytes In      [total, mean]                     4772980, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 948.091µs
Latencies     [min, mean, 50, 90, 95, 99, max]  430.257µs, 882.024µs, 858.456µs, 1.01ms, 1.07ms, 1.394ms, 13.228ms
Bytes In      [total, mean]                     4592971, 153.10
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
Duration      [total, attack, wait]             8m0s, 8m0s, 861.851µs
Latencies     [min, mean, 50, 90, 95, 99, max]  388.906µs, 837.257µs, 832.647µs, 972.75µs, 1.025ms, 1.309ms, 16.112ms
Bytes In      [total, mean]                     7636736, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 880.405µs
Latencies     [min, mean, 50, 90, 95, 99, max]  434.792µs, 859.667µs, 851.452µs, 993.423µs, 1.047ms, 1.318ms, 22.086ms
Bytes In      [total, mean]                     7348877, 153.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 936.791µs
Latencies     [min, mean, 50, 90, 95, 99, max]  434.275µs, 834.34µs, 836.85µs, 955.212µs, 999.745µs, 1.206ms, 4.254ms
Bytes In      [total, mean]                     1909159, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 913.769µs
Latencies     [min, mean, 50, 90, 95, 99, max]  452.248µs, 862.036µs, 859.59µs, 981.982µs, 1.03ms, 1.224ms, 5.425ms
Bytes In      [total, mean]                     1837196, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-oss.png](abrupt-scale-up-affinity-https-oss.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 839.354µs
Latencies     [min, mean, 50, 90, 95, 99, max]  429.134µs, 839.472µs, 844.373µs, 971.135µs, 1.015ms, 1.139ms, 8.115ms
Bytes In      [total, mean]                     1909155, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 939.939µs
Latencies     [min, mean, 50, 90, 95, 99, max]  480.704µs, 870.056µs, 867.661µs, 997.034µs, 1.042ms, 1.192ms, 8.142ms
Bytes In      [total, mean]                     1837196, 153.10
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
Duration      [total, attack, wait]             5m0s, 5m0s, 781µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.747µs, 880.257µs, 862.881µs, 1.016ms, 1.078ms, 1.405ms, 15.565ms
Bytes In      [total, mean]                     4596040, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 753.212µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.949µs, 849.438µs, 840.322µs, 985.358µs, 1.05ms, 1.363ms, 9.304ms
Bytes In      [total, mean]                     4775933, 159.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 1.051ms
Latencies     [min, mean, 50, 90, 95, 99, max]  427.904µs, 869.882µs, 852.202µs, 1.001ms, 1.061ms, 1.348ms, 44.978ms
Bytes In      [total, mean]                     14707047, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.047ms
Latencies     [min, mean, 50, 90, 95, 99, max]  399.655µs, 836.886µs, 826.463µs, 967.339µs, 1.022ms, 1.306ms, 52.641ms
Bytes In      [total, mean]                     15283118, 159.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 799.964µs
Latencies     [min, mean, 50, 90, 95, 99, max]  420.185µs, 816.609µs, 816.423µs, 935.217µs, 979.339µs, 1.205ms, 4.294ms
Bytes In      [total, mean]                     1910348, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 787.365µs
Latencies     [min, mean, 50, 90, 95, 99, max]  448.429µs, 836.05µs, 827.836µs, 956.729µs, 1.004ms, 1.177ms, 5.304ms
Bytes In      [total, mean]                     1838442, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 795.445µs
Latencies     [min, mean, 50, 90, 95, 99, max]  453.2µs, 850.049µs, 837.284µs, 963.627µs, 1.007ms, 1.155ms, 5.583ms
Bytes In      [total, mean]                     1838414, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 841.301µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.29µs, 822.32µs, 814.047µs, 934.883µs, 979.528µs, 1.145ms, 5.637ms
Bytes In      [total, mean]                     1910430, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)
