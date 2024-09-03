# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
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
Duration      [total, attack, wait]             5m0s, 5m0s, 518.975µs
Latencies     [min, mean, 50, 90, 95, 99, max]  426.538µs, 901.008µs, 890.82µs, 1.044ms, 1.105ms, 1.376ms, 11.377ms
Bytes In      [total, mean]                     4626098, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 762.097µs
Latencies     [min, mean, 50, 90, 95, 99, max]  411.265µs, 878.208µs, 872.469µs, 1.024ms, 1.08ms, 1.361ms, 17.026ms
Bytes In      [total, mean]                     4835873, 161.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 888.938µs
Latencies     [min, mean, 50, 90, 95, 99, max]  408.001µs, 875.309µs, 874.083µs, 1.007ms, 1.056ms, 1.295ms, 22.04ms
Bytes In      [total, mean]                     7737488, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 874.716µs
Latencies     [min, mean, 50, 90, 95, 99, max]  451.451µs, 895.987µs, 890.169µs, 1.023ms, 1.071ms, 1.313ms, 22.227ms
Bytes In      [total, mean]                     7401518, 154.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 953.005µs
Latencies     [min, mean, 50, 90, 95, 99, max]  457.648µs, 901.799µs, 895.611µs, 1.049ms, 1.108ms, 1.313ms, 11.714ms
Bytes In      [total, mean]                     1850475, 154.21
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 844.244µs
Latencies     [min, mean, 50, 90, 95, 99, max]  449.379µs, 877.206µs, 874.226µs, 1.017ms, 1.067ms, 1.314ms, 11.747ms
Bytes In      [total, mean]                     1934486, 161.21
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
Duration      [total, attack, wait]             2m0s, 2m0s, 972.192µs
Latencies     [min, mean, 50, 90, 95, 99, max]  454.623µs, 887.707µs, 891.065µs, 1.024ms, 1.07ms, 1.214ms, 33.703ms
Bytes In      [total, mean]                     1934459, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 939.993µs
Latencies     [min, mean, 50, 90, 95, 99, max]  485.897µs, 923.387µs, 918.744µs, 1.067ms, 1.119ms, 1.279ms, 34.539ms
Bytes In      [total, mean]                     1850392, 154.20
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
Duration      [total, attack, wait]             5m0s, 5m0s, 909.014µs
Latencies     [min, mean, 50, 90, 95, 99, max]  460.679µs, 928.486µs, 914.9µs, 1.077ms, 1.14ms, 1.434ms, 13.643ms
Bytes In      [total, mean]                     4625904, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 871.743µs
Latencies     [min, mean, 50, 90, 95, 99, max]  415.562µs, 890.958µs, 882.563µs, 1.042ms, 1.105ms, 1.405ms, 25.819ms
Bytes In      [total, mean]                     4835933, 161.20
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
Duration      [total, attack, wait]             16m0s, 16m0s, 828.444µs
Latencies     [min, mean, 50, 90, 95, 99, max]  416.417µs, 882.164µs, 878.252µs, 1.021ms, 1.074ms, 1.316ms, 43.086ms
Bytes In      [total, mean]                     15475364, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.018ms
Latencies     [min, mean, 50, 90, 95, 99, max]  440.116µs, 919.926µs, 907.162µs, 1.064ms, 1.124ms, 1.37ms, 47.945ms
Bytes In      [total, mean]                     14802983, 154.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.02ms
Latencies     [min, mean, 50, 90, 95, 99, max]  409.065µs, 840.215µs, 841.389µs, 982.195µs, 1.041ms, 1.234ms, 7.107ms
Bytes In      [total, mean]                     1934412, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-plus.png](abrupt-scale-up-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.201ms
Latencies     [min, mean, 50, 90, 95, 99, max]  430.967µs, 895.466µs, 887.337µs, 1.053ms, 1.116ms, 1.307ms, 8.357ms
Bytes In      [total, mean]                     1850365, 154.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 756.544µs
Latencies     [min, mean, 50, 90, 95, 99, max]  448.554µs, 883.048µs, 873.571µs, 1.036ms, 1.097ms, 1.297ms, 39.491ms
Bytes In      [total, mean]                     1850364, 154.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 765.218µs
Latencies     [min, mean, 50, 90, 95, 99, max]  419.144µs, 854.538µs, 846.582µs, 1.008ms, 1.074ms, 1.256ms, 40.114ms
Bytes In      [total, mean]                     1934375, 161.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)
