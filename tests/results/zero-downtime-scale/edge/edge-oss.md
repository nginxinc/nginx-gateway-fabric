# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3a08fdafadfe0fb4a9c25679da1a1fcd6b181474
- Date: 2024-10-15T13:45:52Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1014001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 939.29µs
Latencies     [min, mean, 50, 90, 95, 99, max]  447.092µs, 908.593µs, 889.058µs, 1.037ms, 1.099ms, 1.406ms, 15.858ms
Bytes In      [total, mean]                     4596004, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 901.11µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.052µs, 878.67µs, 865.986µs, 1.006ms, 1.061ms, 1.368ms, 14.501ms
Bytes In      [total, mean]                     4775920, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 913.32µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.02µs, 895.802µs, 889.681µs, 1.028ms, 1.078ms, 1.3ms, 15.706ms
Bytes In      [total, mean]                     7353669, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-oss.png](gradual-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 912.337µs
Latencies     [min, mean, 50, 90, 95, 99, max]  386.177µs, 867.351µs, 865.892µs, 999.816µs, 1.048ms, 1.251ms, 19.417ms
Bytes In      [total, mean]                     7641559, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 847.16µs
Latencies     [min, mean, 50, 90, 95, 99, max]  443.495µs, 896.807µs, 891.053µs, 1.046ms, 1.101ms, 1.288ms, 8.764ms
Bytes In      [total, mean]                     1910359, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 843.736µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.129µs, 911.015µs, 904.124µs, 1.045ms, 1.098ms, 1.309ms, 16.645ms
Bytes In      [total, mean]                     1838400, 153.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 924.596µs
Latencies     [min, mean, 50, 90, 95, 99, max]  417.572µs, 861.287µs, 861.347µs, 992.726µs, 1.036ms, 1.185ms, 11.323ms
Bytes In      [total, mean]                     1910429, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 903.43µs
Latencies     [min, mean, 50, 90, 95, 99, max]  456.415µs, 881.023µs, 879.595µs, 1.007ms, 1.052ms, 1.206ms, 12.289ms
Bytes In      [total, mean]                     1838391, 153.20
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
Duration      [total, attack, wait]             5m0s, 5m0s, 914.573µs
Latencies     [min, mean, 50, 90, 95, 99, max]  455.93µs, 912.022µs, 892.725µs, 1.045ms, 1.107ms, 1.443ms, 23.378ms
Bytes In      [total, mean]                     4602006, 153.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 914.271µs
Latencies     [min, mean, 50, 90, 95, 99, max]  419.145µs, 882.04µs, 868.215µs, 1.009ms, 1.065ms, 1.412ms, 21.119ms
Bytes In      [total, mean]                     4776012, 159.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 824.029µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.835µs, 901.239µs, 888.914µs, 1.028ms, 1.084ms, 1.358ms, 36.675ms
Bytes In      [total, mean]                     14726428, 153.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 957.124µs
Latencies     [min, mean, 50, 90, 95, 99, max]  412.823µs, 869.881µs, 865.519µs, 999.137µs, 1.048ms, 1.307ms, 38.728ms
Bytes In      [total, mean]                     15283183, 159.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 835.687µs
Latencies     [min, mean, 50, 90, 95, 99, max]  444.096µs, 893.078µs, 883.73µs, 1.027ms, 1.081ms, 1.319ms, 13.61ms
Bytes In      [total, mean]                     1840758, 153.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 874.506µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.664µs, 864.186µs, 862.497µs, 988.558µs, 1.035ms, 1.253ms, 13.551ms
Bytes In      [total, mean]                     1910390, 159.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 695.428µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.454µs, 860.36µs, 857.249µs, 1.001ms, 1.055ms, 1.211ms, 13.092ms
Bytes In      [total, mean]                     1910414, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 873.846µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.556µs, 888.246µs, 881.1µs, 1.023ms, 1.076ms, 1.239ms, 23.559ms
Bytes In      [total, mean]                     1840796, 153.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)
