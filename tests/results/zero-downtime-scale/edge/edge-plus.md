# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
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
Duration      [total, attack, wait]             5m0s, 5m0s, 954.535µs
Latencies     [min, mean, 50, 90, 95, 99, max]  415.048µs, 937.174µs, 934.38µs, 1.094ms, 1.152ms, 1.406ms, 12.435ms
Bytes In      [total, mean]                     4836013, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-plus.png](gradual-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 991.205µs
Latencies     [min, mean, 50, 90, 95, 99, max]  479.274µs, 961.844µs, 953.465µs, 1.111ms, 1.169ms, 1.413ms, 12.364ms
Bytes In      [total, mean]                     4655917, 155.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 826.981µs
Latencies     [min, mean, 50, 90, 95, 99, max]  402.249µs, 890.439µs, 887.751µs, 1.035ms, 1.09ms, 1.297ms, 12.06ms
Bytes In      [total, mean]                     7737593, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 930.269µs
Latencies     [min, mean, 50, 90, 95, 99, max]  439.115µs, 927.981µs, 915.944µs, 1.077ms, 1.139ms, 1.348ms, 11.93ms
Bytes In      [total, mean]                     7449595, 155.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 963.179µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.051µs, 896.182µs, 892.984µs, 1.037ms, 1.092ms, 1.268ms, 23.448ms
Bytes In      [total, mean]                     1934369, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-plus.png](abrupt-scale-up-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.025ms
Latencies     [min, mean, 50, 90, 95, 99, max]  485.448µs, 933.597µs, 923.785µs, 1.08ms, 1.14ms, 1.343ms, 10.516ms
Bytes In      [total, mean]                     1862417, 155.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.055ms
Latencies     [min, mean, 50, 90, 95, 99, max]  464.835µs, 963.958µs, 955.367µs, 1.124ms, 1.187ms, 1.34ms, 12.608ms
Bytes In      [total, mean]                     1862303, 155.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-plus.png](abrupt-scale-down-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 976.186µs
Latencies     [min, mean, 50, 90, 95, 99, max]  459.946µs, 943.282µs, 939.805µs, 1.102ms, 1.164ms, 1.318ms, 12.594ms
Bytes In      [total, mean]                     1934388, 161.20
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
Duration      [total, attack, wait]             5m0s, 5m0s, 528.335µs
Latencies     [min, mean, 50, 90, 95, 99, max]  433.12µs, 907.961µs, 903.913µs, 1.054ms, 1.115ms, 1.411ms, 9.334ms
Bytes In      [total, mean]                     4836131, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-plus.png](gradual-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 899.588µs
Latencies     [min, mean, 50, 90, 95, 99, max]  460.25µs, 931.155µs, 918.804µs, 1.068ms, 1.132ms, 1.451ms, 19.026ms
Bytes In      [total, mean]                     4655987, 155.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 902.345µs
Latencies     [min, mean, 50, 90, 95, 99, max]  411.637µs, 851.223µs, 845.28µs, 990.475µs, 1.046ms, 1.289ms, 32.515ms
Bytes In      [total, mean]                     15475283, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.231ms
Latencies     [min, mean, 50, 90, 95, 99, max]  424.17µs, 878.145µs, 867.324µs, 1.02ms, 1.082ms, 1.323ms, 29.739ms
Bytes In      [total, mean]                     14899088, 155.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 930.227µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.024µs, 876.247µs, 872.804µs, 1.017ms, 1.074ms, 1.28ms, 4.556ms
Bytes In      [total, mean]                     1934408, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 909.158µs
Latencies     [min, mean, 50, 90, 95, 99, max]  463.278µs, 915.18µs, 903.478µs, 1.06ms, 1.12ms, 1.307ms, 11.278ms
Bytes In      [total, mean]                     1862379, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 788.663µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.006µs, 851.342µs, 846.807µs, 988.15µs, 1.039ms, 1.201ms, 35.57ms
Bytes In      [total, mean]                     1934488, 161.21
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 885.838µs
Latencies     [min, mean, 50, 90, 95, 99, max]  434.318µs, 883.595µs, 873.371µs, 1.022ms, 1.083ms, 1.258ms, 24.993ms
Bytes In      [total, mean]                     1862363, 155.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)
