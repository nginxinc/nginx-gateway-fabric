# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
- vCPUs per node: 16
- RAM per node: 65853964Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## One NGF Pod runs per node Test Results

### Scale Up Gradually

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 920.846µs
Latencies     [min, mean, 50, 90, 95, 99, max]  411.485µs, 855.405µs, 846.326µs, 992.027µs, 1.055ms, 1.528ms, 12.094ms
Bytes In      [total, mean]                     4856901, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-http-oss.png](gradual-scale-up-affinity-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 835.477µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.379µs, 877.425µs, 865.209µs, 1.015ms, 1.077ms, 1.484ms, 10.128ms
Bytes In      [total, mean]                     4676895, 155.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-affinity-https-oss.png](gradual-scale-up-affinity-https-oss.png)

### Scale Down Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 879.916µs
Latencies     [min, mean, 50, 90, 95, 99, max]  416.857µs, 856.987µs, 853.671µs, 993.963µs, 1.047ms, 1.263ms, 9.923ms
Bytes In      [total, mean]                     7483309, 155.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-https-oss.png](gradual-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         48000, 100.00, 100.00
Duration      [total, attack, wait]             8m0s, 8m0s, 865.816µs
Latencies     [min, mean, 50, 90, 95, 99, max]  406.891µs, 830.509µs, 830.4µs, 966.382µs, 1.017ms, 1.235ms, 10.612ms
Bytes In      [total, mean]                     7771287, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:48000  
Error Set:
```

![gradual-scale-down-affinity-http-oss.png](gradual-scale-down-affinity-http-oss.png)

### Scale Up Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 964.285µs
Latencies     [min, mean, 50, 90, 95, 99, max]  456.767µs, 865.263µs, 860.764µs, 995.816µs, 1.048ms, 1.274ms, 6.511ms
Bytes In      [total, mean]                     1870849, 155.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-https-oss.png](abrupt-scale-up-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 852.484µs
Latencies     [min, mean, 50, 90, 95, 99, max]  406.832µs, 835.797µs, 836.04µs, 966.674µs, 1.015ms, 1.225ms, 5.498ms
Bytes In      [total, mean]                     1942788, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-affinity-http-oss.png](abrupt-scale-up-affinity-http-oss.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 924.091µs
Latencies     [min, mean, 50, 90, 95, 99, max]  415.852µs, 840.745µs, 839.98µs, 973.653µs, 1.019ms, 1.176ms, 6.453ms
Bytes In      [total, mean]                     1870712, 155.89
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-https-oss.png](abrupt-scale-down-affinity-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 766.315µs
Latencies     [min, mean, 50, 90, 95, 99, max]  402.294µs, 811.814µs, 816.225µs, 941.099µs, 985.671µs, 1.133ms, 6.507ms
Bytes In      [total, mean]                     1942795, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-affinity-http-oss.png](abrupt-scale-down-affinity-http-oss.png)

## Multiple NGF Pods run per node Test Results

### Scale Up Gradually

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 910.061µs
Latencies     [min, mean, 50, 90, 95, 99, max]  423.708µs, 875.276µs, 864.898µs, 1.004ms, 1.06ms, 1.376ms, 12.565ms
Bytes In      [total, mean]                     4679934, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

![gradual-scale-up-https-oss.png](gradual-scale-up-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         30000, 100.00, 100.00
Duration      [total, attack, wait]             5m0s, 5m0s, 908.235µs
Latencies     [min, mean, 50, 90, 95, 99, max]  406.086µs, 848.382µs, 841.986µs, 977.23µs, 1.029ms, 1.398ms, 10.725ms
Bytes In      [total, mean]                     4856851, 161.90
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
Duration      [total, attack, wait]             16m0s, 16m0s, 740.733µs
Latencies     [min, mean, 50, 90, 95, 99, max]  410.892µs, 875.413µs, 864.704µs, 1.017ms, 1.086ms, 1.36ms, 20.699ms
Bytes In      [total, mean]                     14975851, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:96000  
Error Set:
```

![gradual-scale-down-https-oss.png](gradual-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         96000, 100.00, 100.00
Duration      [total, attack, wait]             16m0s, 16m0s, 967.509µs
Latencies     [min, mean, 50, 90, 95, 99, max]  400.085µs, 851.909µs, 845.806µs, 995.86µs, 1.062ms, 1.327ms, 12.73ms
Bytes In      [total, mean]                     15542464, 161.90
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
Duration      [total, attack, wait]             2m0s, 2m0s, 964.705µs
Latencies     [min, mean, 50, 90, 95, 99, max]  415.829µs, 850.933µs, 845.823µs, 1.002ms, 1.062ms, 1.283ms, 11.993ms
Bytes In      [total, mean]                     1942775, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-http-oss.png](abrupt-scale-up-http-oss.png)

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 865.341µs
Latencies     [min, mean, 50, 90, 95, 99, max]  413.61µs, 875.028µs, 861.695µs, 1.022ms, 1.089ms, 1.319ms, 12.274ms
Bytes In      [total, mean]                     1872019, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-up-https-oss.png](abrupt-scale-up-https-oss.png)

### Scale Down Abruptly

#### Test: Send https /tea traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 864.336µs
Latencies     [min, mean, 50, 90, 95, 99, max]  446.431µs, 899.507µs, 886.291µs, 1.078ms, 1.154ms, 1.361ms, 7.446ms
Bytes In      [total, mean]                     1872087, 156.01
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-https-oss.png](abrupt-scale-down-https-oss.png)

#### Test: Send http /coffee traffic

```text
Requests      [total, rate, throughput]         12000, 100.01, 100.01
Duration      [total, attack, wait]             2m0s, 2m0s, 897.339µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.501µs, 860.245µs, 853.78µs, 1.033ms, 1.1ms, 1.292ms, 9.213ms
Bytes In      [total, mean]                     1942794, 161.90
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:12000  
Error Set:
```

![abrupt-scale-down-http-oss.png](abrupt-scale-down-http-oss.png)
