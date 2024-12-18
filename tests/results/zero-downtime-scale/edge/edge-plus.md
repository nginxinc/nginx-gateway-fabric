# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 17091ba5d59ca6026f7610e3c2c6200e7ac5cd16
- Date: 2024-12-18T16:52:33Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.6-gke.1125000
- vCPUs per node: 16
- RAM per node: 65853984Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 886.871µs
Latencies     [min, mean, 50, 90, 95, 99, max]  486.53µs, 951.845µs, 940.919µs, 1.098ms, 1.159ms, 1.437ms, 13.574ms
Bytes In      [total, mean]                     4682953, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-plus.png](gradual-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.022ms
Latencies     [min, mean, 50, 90, 95, 99, max]  435.181µs, 929.468µs, 921.983µs, 1.069ms, 1.127ms, 1.425ms, 13.315ms
Bytes In      [total, mean]                     4862869, 162.10
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
Duration      [total, attack, wait]             8m0s, 8m0s, 676.496µs
Latencies     [min, mean, 50, 90, 95, 99, max]  445.563µs, 905.669µs, 900.626µs, 1.045ms, 1.102ms, 1.341ms, 13.894ms
Bytes In      [total, mean]                     7780751, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-plus.png](gradual-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 973.091µs
Latencies     [min, mean, 50, 90, 95, 99, max]  475.296µs, 939.628µs, 926.003µs, 1.084ms, 1.152ms, 1.423ms, 14.54ms
Bytes In      [total, mean]                     7492810, 156.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.106ms
Latencies     [min, mean, 50, 90, 95, 99, max]  474.886µs, 936.833µs, 920.933µs, 1.091ms, 1.166ms, 1.422ms, 12.749ms
Bytes In      [total, mean]                     1873205, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-plus.png](abrupt-scale-up-affinity-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.134ms
Latencies     [min, mean, 50, 90, 95, 99, max]  451.228µs, 898.07µs, 887.98µs, 1.045ms, 1.105ms, 1.361ms, 12.547ms
Bytes In      [total, mean]                     1945169, 162.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 855.057µs
Latencies     [min, mean, 50, 90, 95, 99, max]  432.441µs, 980.678µs, 963.886µs, 1.172ms, 1.262ms, 1.501ms, 43.747ms
Bytes In      [total, mean]                     1945179, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-plus.png](abrupt-scale-down-affinity-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 874.792µs
Latencies     [min, mean, 50, 90, 95, 99, max]  534.115µs, 1.01ms, 986.931µs, 1.198ms, 1.29ms, 1.554ms, 37.461ms
Bytes In      [total, mean]                     1873243, 156.10
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
Duration      [total, attack, wait]             5m0s, 5m0s, 1.032ms
Latencies     [min, mean, 50, 90, 95, 99, max]  436.378µs, 984.194µs, 946.789µs, 1.159ms, 1.285ms, 1.692ms, 17.156ms
Bytes In      [total, mean]                     4683112, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-plus.png](gradual-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.543ms
Latencies     [min, mean, 50, 90, 95, 99, max]  430.165µs, 928.006µs, 909.019µs, 1.089ms, 1.179ms, 1.551ms, 14.35ms
Bytes In      [total, mean]                     4863039, 162.10
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
Duration      [total, attack, wait]             16m0s, 16m0s, 723.686µs
Latencies     [min, mean, 50, 90, 95, 99, max]  432.481µs, 918.039µs, 910.172µs, 1.066ms, 1.133ms, 1.419ms, 14.297ms
Bytes In      [total, mean]                     15561727, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-plus.png](gradual-scale-down-http-plus.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 869.148µs
Latencies     [min, mean, 50, 90, 95, 99, max]  455.261µs, 950.531µs, 935.966µs, 1.099ms, 1.168ms, 1.464ms, 37.876ms
Bytes In      [total, mean]                     14985557, 156.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 999.901µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.039µs, 999.556µs, 950.592µs, 1.297ms, 1.425ms, 1.713ms, 10.475ms
Bytes In      [total, mean]                     1873190, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-plus.png](abrupt-scale-up-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 815.625µs
Latencies     [min, mean, 50, 90, 95, 99, max]  453.334µs, 913.333µs, 883.231µs, 1.144ms, 1.253ms, 1.533ms, 8.265ms
Bytes In      [total, mean]                     1945224, 162.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 911.076µs
Latencies     [min, mean, 50, 90, 95, 99, max]  464.223µs, 911.117µs, 899.875µs, 1.074ms, 1.144ms, 1.322ms, 28.201ms
Bytes In      [total, mean]                     1873182, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-plus.png](abrupt-scale-down-https-plus.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 820.235µs
Latencies     [min, mean, 50, 90, 95, 99, max]  425.145µs, 876.876µs, 871.527µs, 1.046ms, 1.116ms, 1.313ms, 12.53ms
Bytes In      [total, mean]                     1945257, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-plus.png](abrupt-scale-down-http-plus.png)
