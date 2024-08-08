# Results

## Test environment

NGINX Plus: true

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
Duration      [total, attack, wait]             5m0s, 5m0s, 942.809µs
Latencies     [min, mean, 50, 90, 95, 99, max]  458.618µs, 941.816µs, 922.782µs, 1.089ms, 1.154ms, 1.418ms, 23.681ms
Bytes In      [total, mean]                     4563009, 152.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 838.353µs
Latencies     [min, mean, 50, 90, 95, 99, max]  429.355µs, 915.406µs, 903.646µs, 1.06ms, 1.121ms, 1.394ms, 22.686ms
Bytes In      [total, mean]                     4773050, 159.10
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
Duration      [total, attack, wait]             8m0s, 8m0s, 912.161µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.395µs, 902.146µs, 894.005µs, 1.049ms, 1.109ms, 1.33ms, 9.962ms
Bytes In      [total, mean]                     7300716, 152.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 815.068µs
Latencies     [min, mean, 50, 90, 95, 99, max]  415.167µs, 878.754µs, 875.804µs, 1.031ms, 1.087ms, 1.315ms, 8.861ms
Bytes In      [total, mean]                     7636844, 159.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 873.065µs
Latencies     [min, mean, 50, 90, 95, 99, max]  431.169µs, 872.208µs, 866.312µs, 1.006ms, 1.062ms, 1.264ms, 5.729ms
Bytes In      [total, mean]                     1825200, 152.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 880.678µs
Latencies     [min, mean, 50, 90, 95, 99, max]  401.323µs, 872.187µs, 863.423µs, 1.028ms, 1.091ms, 1.275ms, 3.948ms
Bytes In      [total, mean]                     1909245, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.016ms
Latencies     [min, mean, 50, 90, 95, 99, max]  445.96µs, 861.895µs, 865.193µs, 999.949µs, 1.045ms, 1.159ms, 2.535ms
Bytes In      [total, mean]                     1909178, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.055ms
Latencies     [min, mean, 50, 90, 95, 99, max]  438.994µs, 885.169µs, 883.849µs, 1.019ms, 1.063ms, 1.177ms, 24.803ms
Bytes In      [total, mean]                     1825228, 152.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 911.57µs
Latencies     [min, mean, 50, 90, 95, 99, max]  446.131µs, 888.941µs, 879.994µs, 1.03ms, 1.092ms, 1.381ms, 12.021ms
Bytes In      [total, mean]                     4772930, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 797.093µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.351µs, 914.994µs, 900.85µs, 1.059ms, 1.125ms, 1.408ms, 11.915ms
Bytes In      [total, mean]                     4562958, 152.10
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
Duration      [total, attack, wait]             16m0s, 16m0s, 953.992µs
Latencies     [min, mean, 50, 90, 95, 99, max]  434.326µs, 897.924µs, 894.874µs, 1.041ms, 1.095ms, 1.338ms, 11.105ms
Bytes In      [total, mean]                     15273719, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.069ms
Latencies     [min, mean, 50, 90, 95, 99, max]  443.204µs, 931.263µs, 920.396µs, 1.074ms, 1.135ms, 1.371ms, 16.479ms
Bytes In      [total, mean]                     14601648, 152.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 852.38µs
Latencies     [min, mean, 50, 90, 95, 99, max]  458.084µs, 955.606µs, 942.791µs, 1.106ms, 1.168ms, 1.362ms, 12.257ms
Bytes In      [total, mean]                     1825228, 152.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 590.122µs
Latencies     [min, mean, 50, 90, 95, 99, max]  436.387µs, 920.057µs, 915.609µs, 1.086ms, 1.147ms, 1.336ms, 8.788ms
Bytes In      [total, mean]                     1909240, 159.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 789.477µs
Latencies     [min, mean, 50, 90, 95, 99, max]  477.925µs, 932.445µs, 933.377µs, 1.082ms, 1.132ms, 1.29ms, 4.862ms
Bytes In      [total, mean]                     1909211, 159.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 828.368µs
Latencies     [min, mean, 50, 90, 95, 99, max]  497.766µs, 967.155µs, 959.911µs, 1.134ms, 1.193ms, 1.366ms, 6.692ms
Bytes In      [total, mean]                     1825218, 152.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

