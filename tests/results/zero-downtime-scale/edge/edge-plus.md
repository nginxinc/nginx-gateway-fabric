# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: e7d217a8f01fb3c8fc4507ef6f0e7feead667f20
- Date: 2024-11-14T18:42:55Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
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
Duration      [total, attack, wait]             5m0s, 5m0s, 941.193µs
Latencies     [min, mean, 50, 90, 95, 99, max]  408.274µs, 832.805µs, 832.348µs, 963.853µs, 1.015ms, 1.226ms, 12.119ms
Bytes In      [total, mean]                     4836028, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-plus.png](gradual-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 943.297µs
Latencies     [min, mean, 50, 90, 95, 99, max]  413.413µs, 868.848µs, 858.717µs, 997.311µs, 1.055ms, 1.335ms, 15.068ms
Bytes In      [total, mean]                     4655923, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 777.799µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.289µs, 846.567µs, 847.213µs, 974.774µs, 1.021ms, 1.257ms, 16.036ms
Bytes In      [total, mean]                     7737622, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 871.684µs
Latencies     [min, mean, 50, 90, 95, 99, max]  451.158µs, 872.888µs, 867.342µs, 999.583µs, 1.049ms, 1.28ms, 16.856ms
Bytes In      [total, mean]                     7449488, 155.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 839.216µs
Latencies     [min, mean, 50, 90, 95, 99, max]  412.216µs, 827.328µs, 826.882µs, 944.954µs, 986.029µs, 1.157ms, 7.545ms
Bytes In      [total, mean]                     1934359, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 969.121µs
Latencies     [min, mean, 50, 90, 95, 99, max]  467.745µs, 855.826µs, 852.877µs, 976.447µs, 1.022ms, 1.212ms, 6.075ms
Bytes In      [total, mean]                     1862505, 155.21
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.086ms
Latencies     [min, mean, 50, 90, 95, 99, max]  445.748µs, 844.905µs, 841.747µs, 966.834µs, 1.014ms, 1.149ms, 10.252ms
Bytes In      [total, mean]                     1862413, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 977.782µs
Latencies     [min, mean, 50, 90, 95, 99, max]  429.637µs, 820.79µs, 820.371µs, 945.314µs, 990.999µs, 1.119ms, 10.199ms
Bytes In      [total, mean]                     1934426, 161.20
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
Duration      [total, attack, wait]             5m0s, 5m0s, 754.05µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.453µs, 905.139µs, 831.094µs, 960.454µs, 1.011ms, 1.33ms, 1.047s
Bytes In      [total, mean]                     4835964, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 565.701µs
Latencies     [min, mean, 50, 90, 95, 99, max]  455.482µs, 907.551µs, 862.338µs, 996.448µs, 1.053ms, 1.36ms, 1.047s
Bytes In      [total, mean]                     4655923, 155.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 620.297µs
Latencies     [min, mean, 50, 90, 95, 99, max]  405.608µs, 839.322µs, 838.282µs, 965.914µs, 1.013ms, 1.25ms, 23.079ms
Bytes In      [total, mean]                     15475182, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 962.936µs
Latencies     [min, mean, 50, 90, 95, 99, max]  433.619µs, 870.771µs, 863.252µs, 996.003µs, 1.046ms, 1.29ms, 22.949ms
Bytes In      [total, mean]                     14899205, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-plus.png](gradual-scale-down-https-plus.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 866.853µs
Latencies     [min, mean, 50, 90, 95, 99, max]  409.422µs, 841.332µs, 844.856µs, 975.173µs, 1.024ms, 1.182ms, 4.008ms
Bytes In      [total, mean]                     1934371, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 553.714µs
Latencies     [min, mean, 50, 90, 95, 99, max]  460.886µs, 883.007µs, 879.042µs, 1.014ms, 1.067ms, 1.257ms, 8.58ms
Bytes In      [total, mean]                     1862406, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 803.449µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.024µs, 880.184µs, 876.219µs, 1.023ms, 1.072ms, 1.216ms, 6.664ms
Bytes In      [total, mean]                     1862355, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 982.552µs
Latencies     [min, mean, 50, 90, 95, 99, max]  427.658µs, 849.973µs, 848.354µs, 979.91µs, 1.024ms, 1.154ms, 51.405ms
Bytes In      [total, mean]                     1934375, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)
