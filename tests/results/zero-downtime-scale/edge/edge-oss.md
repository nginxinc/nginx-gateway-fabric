# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853980Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 987.169µs
Latencies     [min, mean, 50, 90, 95, 99, max]  428.266µs, 898.492µs, 888.702µs, 1.051ms, 1.109ms, 1.381ms, 12.962ms
Bytes In      [total, mean]                     4865999, 162.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 988.556µs
Latencies     [min, mean, 50, 90, 95, 99, max]  449.333µs, 918.966µs, 905.018µs, 1.077ms, 1.144ms, 1.432ms, 12.423ms
Bytes In      [total, mean]                     4686039, 156.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 853.116µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.889µs, 900.66µs, 893.496µs, 1.053ms, 1.113ms, 1.362ms, 13.105ms
Bytes In      [total, mean]                     7785739, 162.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 959.592µs
Latencies     [min, mean, 50, 90, 95, 99, max]  455.715µs, 934.087µs, 920.517µs, 1.093ms, 1.159ms, 1.428ms, 13.046ms
Bytes In      [total, mean]                     7497528, 156.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.392ms
Latencies     [min, mean, 50, 90, 95, 99, max]  487.674µs, 932.161µs, 924.635µs, 1.088ms, 1.151ms, 1.383ms, 12.482ms
Bytes In      [total, mean]                     1946352, 162.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.256ms
Latencies     [min, mean, 50, 90, 95, 99, max]  486.982µs, 952.759µs, 941.968µs, 1.105ms, 1.167ms, 1.376ms, 5.719ms
Bytes In      [total, mean]                     1874362, 156.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.197ms
Latencies     [min, mean, 50, 90, 95, 99, max]  503.703µs, 983.857µs, 972.529µs, 1.133ms, 1.193ms, 1.36ms, 7.481ms
Bytes In      [total, mean]                     1874292, 156.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.102ms
Latencies     [min, mean, 50, 90, 95, 99, max]  422.367µs, 962.105µs, 954.46µs, 1.116ms, 1.175ms, 1.355ms, 4.928ms
Bytes In      [total, mean]                     1946395, 162.20
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
Duration      [total, attack, wait]             5m0s, 5m0s, 828.186µs
Latencies     [min, mean, 50, 90, 95, 99, max]  463.697µs, 937.232µs, 922.013µs, 1.09ms, 1.156ms, 1.473ms, 14.784ms
Bytes In      [total, mean]                     4685943, 156.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 809.273µs
Latencies     [min, mean, 50, 90, 95, 99, max]  429.94µs, 923.094µs, 911.595µs, 1.082ms, 1.145ms, 1.448ms, 15.849ms
Bytes In      [total, mean]                     4866096, 162.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 879.282µs
Latencies     [min, mean, 50, 90, 95, 99, max]  429.627µs, 938.31µs, 925.679µs, 1.095ms, 1.159ms, 1.438ms, 21.727ms
Bytes In      [total, mean]                     14995043, 156.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 769.451µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.371µs, 915.1µs, 905.553µs, 1.071ms, 1.135ms, 1.42ms, 16.634ms
Bytes In      [total, mean]                     15571090, 162.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-oss.png](gradual-scale-down-http-oss.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 763.563µs
Latencies     [min, mean, 50, 90, 95, 99, max]  461.763µs, 922.722µs, 906.525µs, 1.07ms, 1.134ms, 1.466ms, 13.287ms
Bytes In      [total, mean]                     1874421, 156.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 828.463µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.692µs, 890.963µs, 884.555µs, 1.045ms, 1.106ms, 1.393ms, 12.309ms
Bytes In      [total, mean]                     1946381, 162.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.107ms
Latencies     [min, mean, 50, 90, 95, 99, max]  458.875µs, 959.177µs, 953.252µs, 1.14ms, 1.199ms, 1.365ms, 10.216ms
Bytes In      [total, mean]                     1946462, 162.21
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 977.104µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.553µs, 985.637µs, 973.975µs, 1.171ms, 1.24ms, 1.428ms, 7.298ms
Bytes In      [total, mean]                     1874451, 156.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)
