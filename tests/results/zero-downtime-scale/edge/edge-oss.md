# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 809c0838e2f2658c3c4cd48325ffb0bc5a92a002
- Date: 2024-08-08T18:03:35Z
- Dirty: false

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
Duration      [total, attack, wait]             5m0s, 5m0s, 1.223ms
Latencies     [min, mean, 50, 90, 95, 99, max]  445.183µs, 977.19µs, 926.748µs, 1.103ms, 1.177ms, 1.424ms, 1.005s
Bytes In      [total, mean]                     4592956, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https.png](gradual-scale-up-affinity-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.099ms
Latencies     [min, mean, 50, 90, 95, 99, max]  440.581µs, 905.167µs, 894.329µs, 1.061ms, 1.129ms, 1.403ms, 23.88ms
Bytes In      [total, mean]                     4802940, 160.10
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
Duration      [total, attack, wait]             8m0s, 8m0s, 952.971µs
Latencies     [min, mean, 50, 90, 95, 99, max]  439.56µs, 914.924µs, 912.558µs, 1.059ms, 1.115ms, 1.343ms, 12.102ms
Bytes In      [total, mean]                     7684727, 160.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http.png](gradual-scale-down-affinity-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 601.881µs
Latencies     [min, mean, 50, 90, 95, 99, max]  486.48µs, 945.177µs, 936.532µs, 1.086ms, 1.144ms, 1.367ms, 12.823ms
Bytes In      [total, mean]                     7348717, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https.png](gradual-scale-down-affinity-https.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 798.198µs
Latencies     [min, mean, 50, 90, 95, 99, max]  490.902µs, 956.137µs, 943.75µs, 1.119ms, 1.189ms, 1.389ms, 10.008ms
Bytes In      [total, mean]                     1837153, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https.png](abrupt-scale-up-affinity-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 941.827µs
Latencies     [min, mean, 50, 90, 95, 99, max]  465.874µs, 896.849µs, 898.529µs, 1.037ms, 1.082ms, 1.237ms, 9.993ms
Bytes In      [total, mean]                     1921242, 160.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http.png](abrupt-scale-up-affinity-http.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 827.899µs
Latencies     [min, mean, 50, 90, 95, 99, max]  489.648µs, 939.147µs, 936.079µs, 1.09ms, 1.148ms, 1.329ms, 11.684ms
Bytes In      [total, mean]                     1837157, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https.png](abrupt-scale-down-affinity-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 811.537µs
Latencies     [min, mean, 50, 90, 95, 99, max]  450.599µs, 903.375µs, 911.36µs, 1.048ms, 1.097ms, 1.231ms, 4.209ms
Bytes In      [total, mean]                     1921188, 160.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http.png](abrupt-scale-down-affinity-http.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 911.675µs
Latencies     [min, mean, 50, 90, 95, 99, max]  461.987µs, 978.789µs, 961.781µs, 1.15ms, 1.225ms, 1.497ms, 19.815ms
Bytes In      [total, mean]                     4593066, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https.png](gradual-scale-up-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 1.136ms
Latencies     [min, mean, 50, 90, 95, 99, max]  444.725µs, 928.921µs, 921.076µs, 1.089ms, 1.157ms, 1.441ms, 20.337ms
Bytes In      [total, mean]                     4803008, 160.10
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
Duration      [total, attack, wait]             16m0s, 16m0s, 620.326µs
Latencies     [min, mean, 50, 90, 95, 99, max]  441.037µs, 926.771µs, 920.769µs, 1.089ms, 1.155ms, 1.389ms, 11.721ms
Bytes In      [total, mean]                     15369650, 160.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-http.png](gradual-scale-down-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 952.832µs
Latencies     [min, mean, 50, 90, 95, 99, max]  438.206µs, 966.22µs, 952.216µs, 1.143ms, 1.219ms, 1.439ms, 12.852ms
Bytes In      [total, mean]                     14697549, 153.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 1.119ms
Latencies     [min, mean, 50, 90, 95, 99, max]  447.445µs, 940.825µs, 933.705µs, 1.108ms, 1.176ms, 1.412ms, 11.945ms
Bytes In      [total, mean]                     1921184, 160.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http.png](abrupt-scale-up-http.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 932.16µs
Latencies     [min, mean, 50, 90, 95, 99, max]  452.739µs, 955.606µs, 940.67µs, 1.138ms, 1.212ms, 1.427ms, 8.525ms
Bytes In      [total, mean]                     1837194, 153.10
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
Duration      [total, attack, wait]             2m0s, 2m0s, 981.584µs
Latencies     [min, mean, 50, 90, 95, 99, max]  501.041µs, 993.484µs, 993.087µs, 1.165ms, 1.226ms, 1.391ms, 7.293ms
Bytes In      [total, mean]                     1837233, 153.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https.png](abrupt-scale-down-https.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 1.071ms
Latencies     [min, mean, 50, 90, 95, 99, max]  447.174µs, 964.502µs, 967.597µs, 1.137ms, 1.198ms, 1.372ms, 3.44ms
Bytes In      [total, mean]                     1921198, 160.10
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http.png](abrupt-scale-down-http.png)

