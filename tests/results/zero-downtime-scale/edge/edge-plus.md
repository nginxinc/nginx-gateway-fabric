# Results

## Test environment

NGINX Plus: true

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
Duration      [total, attack, wait]             5m0s, 5m0s, 702.389µs
Latencies     [min, mean, 50, 90, 95, 99, max]  400.298µs, 889.573µs, 877.905µs, 1.039ms, 1.099ms, 1.429ms, 22.696ms
Bytes In      [total, mean]                     4806000, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-plus.png](gradual-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 858.769µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.235µs, 932.766µs, 918.529µs, 1.084ms, 1.148ms, 1.451ms, 21.927ms
Bytes In      [total, mean]                     4626075, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 995.623µs
Latencies     [min, mean, 50, 90, 95, 99, max]  431.898µs, 931.625µs, 916.287µs, 1.099ms, 1.175ms, 1.415ms, 23.236ms
Bytes In      [total, mean]                     7401576, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-plus.png](gradual-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 1.021ms
Latencies     [min, mean, 50, 90, 95, 99, max]  394.896µs, 910.879µs, 896.82µs, 1.08ms, 1.153ms, 1.422ms, 27.642ms
Bytes In      [total, mean]                     7689695, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 936.364µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.959µs, 888.23µs, 887.429µs, 1.036ms, 1.087ms, 1.323ms, 10.206ms
Bytes In      [total, mean]                     1922443, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 776.401µs
Latencies     [min, mean, 50, 90, 95, 99, max]  447.882µs, 930.841µs, 917.586µs, 1.104ms, 1.176ms, 1.374ms, 12.755ms
Bytes In      [total, mean]                     1850355, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 817.595µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.569µs, 908.306µs, 905.877µs, 1.051ms, 1.101ms, 1.236ms, 27.237ms
Bytes In      [total, mean]                     1850442, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 838.215µs
Latencies     [min, mean, 50, 90, 95, 99, max]  411.876µs, 882.203µs, 883.378µs, 1.026ms, 1.074ms, 1.206ms, 13.45ms
Bytes In      [total, mean]                     1922381, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 876.661µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.48µs, 924.895µs, 909.414µs, 1.089ms, 1.161ms, 1.457ms, 11.21ms
Bytes In      [total, mean]                     4625890, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 557.105µs
Latencies     [min, mean, 50, 90, 95, 99, max]  405.998µs, 928.718µs, 914.734µs, 1.113ms, 1.185ms, 1.48ms, 13.894ms
Bytes In      [total, mean]                     4806032, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 976.669µs
Latencies     [min, mean, 50, 90, 95, 99, max]  405.181µs, 903.273µs, 898.013µs, 1.059ms, 1.12ms, 1.363ms, 23.914ms
Bytes In      [total, mean]                     15379200, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 994.86µs
Latencies     [min, mean, 50, 90, 95, 99, max]  416.117µs, 936.093µs, 926.338µs, 1.097ms, 1.161ms, 1.416ms, 13.766ms
Bytes In      [total, mean]                     14803203, 154.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 908.763µs
Latencies     [min, mean, 50, 90, 95, 99, max]  470.723µs, 945.029µs, 933.242µs, 1.107ms, 1.17ms, 1.441ms, 12.164ms
Bytes In      [total, mean]                     1850328, 154.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1ms
Latencies     [min, mean, 50, 90, 95, 99, max]  433.61µs, 896.602µs, 889.18µs, 1.056ms, 1.118ms, 1.303ms, 21.715ms
Bytes In      [total, mean]                     1922326, 160.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 994.561µs
Latencies     [min, mean, 50, 90, 95, 99, max]  474.453µs, 929.685µs, 921.959µs, 1.1ms, 1.163ms, 1.327ms, 18.075ms
Bytes In      [total, mean]                     1850404, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.016ms
Latencies     [min, mean, 50, 90, 95, 99, max]  446.404µs, 898.613µs, 898.444µs, 1.06ms, 1.117ms, 1.266ms, 6.258ms
Bytes In      [total, mean]                     1922442, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)
