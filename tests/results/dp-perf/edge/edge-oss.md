# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 929413c15af7bee3adb32e103c9d1513a693da16
- Date: 2024-11-28T12:52:45Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853964Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.97
Duration      [total, attack, wait]             30s, 29.999s, 919.839µs
Latencies     [min, mean, 50, 90, 95, 99, max]  555.081µs, 763.733µs, 741.7µs, 856.388µs, 905.52µs, 1.037ms, 12.058ms
Bytes In      [total, mean]                     4799840, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 1.001ms
Latencies     [min, mean, 50, 90, 95, 99, max]  571.797µs, 799.694µs, 781.424µs, 902.428µs, 948.675µs, 1.072ms, 12.582ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 798.719µs
Latencies     [min, mean, 50, 90, 95, 99, max]  593.311µs, 810.059µs, 794.827µs, 916.869µs, 967.97µs, 1.102ms, 12.056ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 767.537µs
Latencies     [min, mean, 50, 90, 95, 99, max]  570.611µs, 793.224µs, 777.61µs, 895.269µs, 944.58µs, 1.088ms, 9.078ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 30s, 822.686µs
Latencies     [min, mean, 50, 90, 95, 99, max]  578.767µs, 799.921µs, 782.94µs, 907.868µs, 958.671µs, 1.101ms, 13.291ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
