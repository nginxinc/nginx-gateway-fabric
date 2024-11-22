# Results

## Test environment

NGINX Plus: false

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

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 674.073µs
Latencies     [min, mean, 50, 90, 95, 99, max]  419.791µs, 840.72µs, 834.34µs, 962.098µs, 1.015ms, 1.294ms, 10.169ms
Bytes In      [total, mean]                     4679929, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 877.124µs
Latencies     [min, mean, 50, 90, 95, 99, max]  412.325µs, 829.237µs, 822.155µs, 943.97µs, 993.732µs, 1.323ms, 12.09ms
Bytes In      [total, mean]                     4857151, 161.91
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
Duration      [total, attack, wait]             8m0s, 8m0s, 867.199µs
Latencies     [min, mean, 50, 90, 95, 99, max]  418.599µs, 852.891µs, 843.777µs, 979.471µs, 1.035ms, 1.33ms, 14.512ms
Bytes In      [total, mean]                     7488117, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-oss.png](gradual-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 886.292µs
Latencies     [min, mean, 50, 90, 95, 99, max]  405.98µs, 825.969µs, 821.489µs, 947.564µs, 997.377µs, 1.274ms, 15.137ms
Bytes In      [total, mean]                     7771168, 161.90
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
Duration      [total, attack, wait]             2m0s, 2m0s, 969.83µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.695µs, 865.281µs, 855.081µs, 992.225µs, 1.055ms, 1.414ms, 12.639ms
Bytes In      [total, mean]                     1942752, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 774.619µs
Latencies     [min, mean, 50, 90, 95, 99, max]  425.067µs, 879.019µs, 872.194µs, 1.013ms, 1.078ms, 1.447ms, 7.09ms
Bytes In      [total, mean]                     1872026, 156.00
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
Duration      [total, attack, wait]             2m0s, 2m0s, 791.319µs
Latencies     [min, mean, 50, 90, 95, 99, max]  397.317µs, 840.677µs, 845.387µs, 967.105µs, 1.013ms, 1.198ms, 9.9ms
Bytes In      [total, mean]                     1942785, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 810.08µs
Latencies     [min, mean, 50, 90, 95, 99, max]  475.946µs, 868.195µs, 867.184µs, 991.184µs, 1.041ms, 1.223ms, 5.819ms
Bytes In      [total, mean]                     1871908, 155.99
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
Duration      [total, attack, wait]             5m0s, 5m0s, 756.924µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.574µs, 859.757µs, 839.255µs, 970.782µs, 1.03ms, 1.433ms, 19.967ms
Bytes In      [total, mean]                     4682982, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 875.242µs
Latencies     [min, mean, 50, 90, 95, 99, max]  390.062µs, 823.832µs, 812.588µs, 941.574µs, 993.373µs, 1.379ms, 13.509ms
Bytes In      [total, mean]                     4862819, 162.09
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
Duration      [total, attack, wait]             16m0s, 16m0s, 810.664µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.774µs, 864.05µs, 853.331µs, 978.297µs, 1.03ms, 1.36ms, 52.309ms
Bytes In      [total, mean]                     14985561, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 911.775µs
Latencies     [min, mean, 50, 90, 95, 99, max]  380.949µs, 837.707µs, 832.349µs, 953.169µs, 1.001ms, 1.32ms, 50.143ms
Bytes In      [total, mean]                     15561752, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http-oss.png](gradual-scale-down-http-oss.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 923.877µs
Latencies     [min, mean, 50, 90, 95, 99, max]  421.746µs, 832.719µs, 831.849µs, 952.303µs, 999.227µs, 1.298ms, 10.711ms
Bytes In      [total, mean]                     1945216, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 923.761µs
Latencies     [min, mean, 50, 90, 95, 99, max]  435.295µs, 857.491µs, 852.757µs, 975.748µs, 1.023ms, 1.305ms, 7.941ms
Bytes In      [total, mean]                     1873208, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 844.211µs
Latencies     [min, mean, 50, 90, 95, 99, max]  440.009µs, 856.225µs, 860.69µs, 981.975µs, 1.025ms, 1.195ms, 13.743ms
Bytes In      [total, mean]                     1945223, 162.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 868.665µs
Latencies     [min, mean, 50, 90, 95, 99, max]  458.08µs, 884.664µs, 881.714µs, 1.006ms, 1.051ms, 1.249ms, 13.184ms
Bytes In      [total, mean]                     1873180, 156.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)
