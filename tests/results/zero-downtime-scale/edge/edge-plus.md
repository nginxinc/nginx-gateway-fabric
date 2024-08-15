# Results

## Test environment

NGINX Plus: true

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

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 952.729µs
Latencies     [min, mean, 50, 90, 95, 99, max]  454.567µs, 940.698µs, 931.277µs, 1.078ms, 1.136ms, 1.357ms, 19.577ms
Bytes In      [total, mean]                     4587013, 152.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 767.806µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.125µs, 908.744µs, 907.09µs, 1.047ms, 1.098ms, 1.364ms, 14.861ms
Bytes In      [total, mean]                     4791034, 159.70
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
Duration      [total, attack, wait]             8m0s, 8m0s, 689.825µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.82µs, 888.906µs, 889.49µs, 1.028ms, 1.078ms, 1.347ms, 20.196ms
Bytes In      [total, mean]                     7665470, 159.70
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 744.752µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.067µs, 912.33µs, 907.634µs, 1.049ms, 1.102ms, 1.347ms, 20.464ms
Bytes In      [total, mean]                     7339080, 152.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.148ms
Latencies     [min, mean, 50, 90, 95, 99, max]  460.712µs, 882.499µs, 880.208µs, 1.02ms, 1.068ms, 1.244ms, 9.074ms
Bytes In      [total, mean]                     1834680, 152.89
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.18ms
Latencies     [min, mean, 50, 90, 95, 99, max]  428.344µs, 862.118µs, 862.208µs, 1.007ms, 1.052ms, 1.236ms, 9.083ms
Bytes In      [total, mean]                     1916474, 159.71
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.027ms
Latencies     [min, mean, 50, 90, 95, 99, max]  464.838µs, 918.721µs, 919.978µs, 1.06ms, 1.105ms, 1.244ms, 11.629ms
Bytes In      [total, mean]                     1834731, 152.89
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.039ms
Latencies     [min, mean, 50, 90, 95, 99, max]  455.228µs, 899.607µs, 903.219µs, 1.042ms, 1.085ms, 1.223ms, 11.288ms
Bytes In      [total, mean]                     1916412, 159.70
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 557.058µs
Latencies     [min, mean, 50, 90, 95, 99, max]  425.666µs, 881.25µs, 876.97µs, 1.019ms, 1.073ms, 1.374ms, 16.057ms
Bytes In      [total, mean]                     4799925, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 829.03µs
Latencies     [min, mean, 50, 90, 95, 99, max]  452.719µs, 908.347µs, 897.76µs, 1.047ms, 1.108ms, 1.409ms, 10.825ms
Bytes In      [total, mean]                     4590033, 153.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 965.681µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.359µs, 891.388µs, 887.74µs, 1.027ms, 1.079ms, 1.331ms, 28.941ms
Bytes In      [total, mean]                     15360059, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 930.193µs
Latencies     [min, mean, 50, 90, 95, 99, max]  402.803µs, 917.57µs, 909.154µs, 1.052ms, 1.107ms, 1.361ms, 23.221ms
Bytes In      [total, mean]                     14687799, 153.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 934.068µs
Latencies     [min, mean, 50, 90, 95, 99, max]  485.389µs, 932.491µs, 923.049µs, 1.07ms, 1.127ms, 1.364ms, 12.554ms
Bytes In      [total, mean]                     1835959, 153.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.904ms
Latencies     [min, mean, 50, 90, 95, 99, max]  467.756µs, 905.39µs, 905.54µs, 1.05ms, 1.101ms, 1.324ms, 10.761ms
Bytes In      [total, mean]                     1920019, 160.00
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.057ms
Latencies     [min, mean, 50, 90, 95, 99, max]  492.06µs, 941.087µs, 938.497µs, 1.101ms, 1.159ms, 1.311ms, 9.64ms
Bytes In      [total, mean]                     1920055, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.057ms
Latencies     [min, mean, 50, 90, 95, 99, max]  465.962µs, 958.715µs, 951.081µs, 1.123ms, 1.185ms, 1.35ms, 9.641ms
Bytes In      [total, mean]                     1835946, 153.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

