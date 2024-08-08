# Results

## Test environment

NGINX Plus: false

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
Duration      [total, attack, wait]             5m0s, 5m0s, 900.46µs
Latencies     [min, mean, 50, 90, 95, 99, max]  457.981µs, 908.335µs, 895.192µs, 1.043ms, 1.105ms, 1.368ms, 12.229ms
Bytes In      [total, mean]                     4625932, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https.png](gradual-scale-up-affinity-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 862.021µs
Latencies     [min, mean, 50, 90, 95, 99, max]  405.395µs, 868.889µs, 864.535µs, 1.004ms, 1.061ms, 1.343ms, 13.286ms
Bytes In      [total, mean]                     4835871, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http.png](gradual-scale-up-affinity-http.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 896.917µs
Latencies     [min, mean, 50, 90, 95, 99, max]  404.132µs, 870.451µs, 867.853µs, 1.01ms, 1.063ms, 1.293ms, 10.241ms
Bytes In      [total, mean]                     7737648, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http.png](gradual-scale-down-affinity-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 919.699µs
Latencies     [min, mean, 50, 90, 95, 99, max]  431.902µs, 890.121µs, 883.321µs, 1.028ms, 1.083ms, 1.293ms, 16.637ms
Bytes In      [total, mean]                     7401490, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https.png](gradual-scale-down-affinity-https.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 906.482µs
Latencies     [min, mean, 50, 90, 95, 99, max]  391.212µs, 864.148µs, 867.788µs, 1.006ms, 1.054ms, 1.227ms, 4.459ms
Bytes In      [total, mean]                     1934424, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http.png](abrupt-scale-up-affinity-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 970.733µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.813µs, 899.858µs, 898.925µs, 1.044ms, 1.097ms, 1.277ms, 5.826ms
Bytes In      [total, mean]                     1850355, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https.png](abrupt-scale-up-affinity-https.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 731.031µs
Latencies     [min, mean, 50, 90, 95, 99, max]  439.936µs, 902.663µs, 902.24µs, 1.052ms, 1.106ms, 1.264ms, 11.088ms
Bytes In      [total, mean]                     1850329, 154.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https.png](abrupt-scale-down-affinity-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 910.291µs
Latencies     [min, mean, 50, 90, 95, 99, max]  441.434µs, 881.572µs, 885.044µs, 1.028ms, 1.075ms, 1.218ms, 7.272ms
Bytes In      [total, mean]                     1934390, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http.png](abrupt-scale-down-affinity-http.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 782.764µs
Latencies     [min, mean, 50, 90, 95, 99, max]  399.871µs, 877.754µs, 871.563µs, 1.025ms, 1.086ms, 1.358ms, 11.094ms
Bytes In      [total, mean]                     4836077, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http.png](gradual-scale-up-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 992.791µs
Latencies     [min, mean, 50, 90, 95, 99, max]  439.443µs, 904.22µs, 894.202µs, 1.045ms, 1.109ms, 1.385ms, 10.086ms
Bytes In      [total, mean]                     4625989, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https.png](gradual-scale-up-https.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 579.411µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.189µs, 909.96µs, 902.119µs, 1.062ms, 1.124ms, 1.394ms, 33.289ms
Bytes In      [total, mean]                     14803363, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https.png](gradual-scale-down-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 847.127µs
Latencies     [min, mean, 50, 90, 95, 99, max]  396.604µs, 888.417µs, 883.884µs, 1.039ms, 1.099ms, 1.367ms, 35.413ms
Bytes In      [total, mean]                     15475006, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http.png](gradual-scale-down-http.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 804.774µs
Latencies     [min, mean, 50, 90, 95, 99, max]  398.813µs, 900.553µs, 895.332µs, 1.054ms, 1.121ms, 1.42ms, 9.106ms
Bytes In      [total, mean]                     1934410, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http.png](abrupt-scale-up-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 852.018µs
Latencies     [min, mean, 50, 90, 95, 99, max]  487.357µs, 946.866µs, 932.744µs, 1.103ms, 1.169ms, 1.448ms, 12.626ms
Bytes In      [total, mean]                     1850442, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https.png](abrupt-scale-up-https.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.017ms
Latencies     [min, mean, 50, 90, 95, 99, max]  414.609µs, 922.025µs, 918.849µs, 1.091ms, 1.153ms, 1.367ms, 6.922ms
Bytes In      [total, mean]                     1934419, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http.png](abrupt-scale-down-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.023ms
Latencies     [min, mean, 50, 90, 95, 99, max]  472.249µs, 949.451µs, 940.979µs, 1.116ms, 1.185ms, 1.424ms, 10.447ms
Bytes In      [total, mean]                     1850401, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https.png](abrupt-scale-down-https.png)

