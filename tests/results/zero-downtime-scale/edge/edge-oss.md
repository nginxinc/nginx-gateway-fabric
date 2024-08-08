# Results

## Test environment

NGINX Plus: false

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
Duration      [total, attack, wait]             5m0s, 5m0s, 904.888µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.23µs, 913.513µs, 889.019µs, 1.064ms, 1.142ms, 1.401ms, 12.652ms
Bytes In      [total, mean]                     4595909, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https.png](gradual-scale-up-affinity-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 904.783µs
Latencies     [min, mean, 50, 90, 95, 99, max]  441.813µs, 887.312µs, 869.05µs, 1.026ms, 1.096ms, 1.379ms, 12.48ms
Bytes In      [total, mean]                     4805926, 160.20
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
Duration      [total, attack, wait]             8m0s, 8m0s, 663.831µs
Latencies     [min, mean, 50, 90, 95, 99, max]  413.061µs, 858.199µs, 846.865µs, 1.013ms, 1.081ms, 1.303ms, 14.069ms
Bytes In      [total, mean]                     7689583, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http.png](gradual-scale-down-affinity-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 1.153ms
Latencies     [min, mean, 50, 90, 95, 99, max]  440.462µs, 884.376µs, 866.673µs, 1.04ms, 1.113ms, 1.344ms, 13.788ms
Bytes In      [total, mean]                     7353556, 153.20
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
Duration      [total, attack, wait]             2m0s, 2m0s, 748.826µs
Latencies     [min, mean, 50, 90, 95, 99, max]  414.935µs, 876.97µs, 862.023µs, 1.04ms, 1.109ms, 1.339ms, 6.467ms
Bytes In      [total, mean]                     1922356, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http.png](abrupt-scale-up-affinity-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 853.537µs
Latencies     [min, mean, 50, 90, 95, 99, max]  475.862µs, 923.434µs, 903.004µs, 1.104ms, 1.178ms, 1.394ms, 16.292ms
Bytes In      [total, mean]                     1838428, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https.png](abrupt-scale-up-affinity-https.png)

### Scale Down Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.00
Duration      [total, attack, wait]             2m0s, 2m0s, 6.943ms
Latencies     [min, mean, 50, 90, 95, 99, max]  445.682µs, 940.34µs, 926.877µs, 1.127ms, 1.203ms, 1.374ms, 13.686ms
Bytes In      [total, mean]                     1922371, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http.png](abrupt-scale-down-affinity-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.02ms
Latencies     [min, mean, 50, 90, 95, 99, max]  467.274µs, 973.76µs, 950.271µs, 1.175ms, 1.254ms, 1.439ms, 23.552ms
Bytes In      [total, mean]                     1838474, 153.21
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https.png](abrupt-scale-down-affinity-https.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 757.853µs
Latencies     [min, mean, 50, 90, 95, 99, max]  428.942µs, 881.603µs, 866.883µs, 1.022ms, 1.084ms, 1.356ms, 12.869ms
Bytes In      [total, mean]                     4595921, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https.png](gradual-scale-up-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 779.771µs
Latencies     [min, mean, 50, 90, 95, 99, max]  424.286µs, 855.226µs, 847.3µs, 999.364µs, 1.058ms, 1.323ms, 10.769ms
Bytes In      [total, mean]                     4806006, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-http.png](gradual-scale-up-http.png)

### Scale Down Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 807.454µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.299µs, 857.717µs, 849.208µs, 998.356µs, 1.055ms, 1.296ms, 25.564ms
Bytes In      [total, mean]                     15379105, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http.png](gradual-scale-down-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 802.146µs
Latencies     [min, mean, 50, 90, 95, 99, max]  430.568µs, 883.542µs, 871.068µs, 1.024ms, 1.084ms, 1.332ms, 31.788ms
Bytes In      [total, mean]                     14707312, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https.png](gradual-scale-down-https.png)

### Scale Up Abruptly

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 797.694µs
Latencies     [min, mean, 50, 90, 95, 99, max]  436.054µs, 852.464µs, 843.609µs, 993.392µs, 1.049ms, 1.259ms, 10.488ms
Bytes In      [total, mean]                     1922444, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http.png](abrupt-scale-up-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 912.691µs
Latencies     [min, mean, 50, 90, 95, 99, max]  455.927µs, 874.227µs, 859.306µs, 1.009ms, 1.075ms, 1.33ms, 11.52ms
Bytes In      [total, mean]                     1838359, 153.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https.png](abrupt-scale-up-https.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 979.206µs
Latencies     [min, mean, 50, 90, 95, 99, max]  464.331µs, 877.352µs, 861.432µs, 999.62µs, 1.049ms, 1.191ms, 10.085ms
Bytes In      [total, mean]                     1838503, 153.21
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https.png](abrupt-scale-down-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 978.295µs
Latencies     [min, mean, 50, 90, 95, 99, max]  437.909µs, 851.952µs, 839.434µs, 973.315µs, 1.02ms, 1.163ms, 7.785ms
Bytes In      [total, mean]                     1922408, 160.20
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http.png](abrupt-scale-down-http.png)
