# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 9a85dbcc0797e31557a3731688795aa166ee0f96
- Date: 2024-08-13T21:12:05Z
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
Duration      [total, attack, wait]             5m0s, 5m0s, 1.114ms
Latencies     [min, mean, 50, 90, 95, 99, max]  449.75µs, 911.702µs, 900.385µs, 1.019ms, 1.069ms, 1.344ms, 13.752ms
Bytes In      [total, mean]                     4565908, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.017ms
Latencies     [min, mean, 50, 90, 95, 99, max]  449.683µs, 887.251µs, 881.387µs, 998.312µs, 1.045ms, 1.365ms, 14.096ms
Bytes In      [total, mean]                     4776008, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 926.769µs
Latencies     [min, mean, 50, 90, 95, 99, max]  466.912µs, 924.764µs, 923.218µs, 1.05ms, 1.098ms, 1.304ms, 12.057ms
Bytes In      [total, mean]                     7641701, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 931.87µs
Latencies     [min, mean, 50, 90, 95, 99, max]  487.371µs, 951.705µs, 941.325µs, 1.075ms, 1.132ms, 1.335ms, 12.607ms
Bytes In      [total, mean]                     7305525, 152.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.05ms
Latencies     [min, mean, 50, 90, 95, 99, max]  486.019µs, 944.509µs, 938.484µs, 1.082ms, 1.138ms, 1.336ms, 4.334ms
Bytes In      [total, mean]                     1910319, 159.19
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.05ms
Latencies     [min, mean, 50, 90, 95, 99, max]  545.279µs, 972.604µs, 955.006µs, 1.107ms, 1.172ms, 1.386ms, 11.678ms
Bytes In      [total, mean]                     1826411, 152.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 956.441µs
Latencies     [min, mean, 50, 90, 95, 99, max]  529.38µs, 984.48µs, 971.856µs, 1.097ms, 1.144ms, 1.28ms, 6.422ms
Bytes In      [total, mean]                     1826403, 152.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 991.231µs
Latencies     [min, mean, 50, 90, 95, 99, max]  538.784µs, 960.249µs, 952.878µs, 1.074ms, 1.117ms, 1.237ms, 4.488ms
Bytes In      [total, mean]                     1910450, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 982.805µs
Latencies     [min, mean, 50, 90, 95, 99, max]  455.338µs, 905.599µs, 904.267µs, 1.019ms, 1.066ms, 1.369ms, 7.871ms
Bytes In      [total, mean]                     4776013, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http-oss.png](gradual-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 806.041µs
Latencies     [min, mean, 50, 90, 95, 99, max]  476.73µs, 922.222µs, 915.298µs, 1.032ms, 1.082ms, 1.375ms, 9.75ms
Bytes In      [total, mean]                     4572028, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 1.007ms
Latencies     [min, mean, 50, 90, 95, 99, max]  445.865µs, 887.204µs, 880.343µs, 1.006ms, 1.055ms, 1.319ms, 37.472ms
Bytes In      [total, mean]                     14630301, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 977.391µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.218µs, 862.403µs, 859.832µs, 984.84µs, 1.031ms, 1.299ms, 31.685ms
Bytes In      [total, mean]                     15282845, 159.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 829.772µs
Latencies     [min, mean, 50, 90, 95, 99, max]  443.409µs, 864.3µs, 859.901µs, 986.131µs, 1.037ms, 1.272ms, 10.525ms
Bytes In      [total, mean]                     1828812, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 637.057µs
Latencies     [min, mean, 50, 90, 95, 99, max]  436.37µs, 831.689µs, 832.971µs, 954.284µs, 1.001ms, 1.202ms, 4.742ms
Bytes In      [total, mean]                     1910384, 159.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 926.788µs
Latencies     [min, mean, 50, 90, 95, 99, max]  442.954µs, 860.614µs, 860.834µs, 985.004µs, 1.033ms, 1.175ms, 12.153ms
Bytes In      [total, mean]                     1910368, 159.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 923.169µs
Latencies     [min, mean, 50, 90, 95, 99, max]  468.689µs, 879.118µs, 878.701µs, 1.003ms, 1.051ms, 1.191ms, 12.165ms
Bytes In      [total, mean]                     1828806, 152.40
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)

