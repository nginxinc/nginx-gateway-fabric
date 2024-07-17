# Results

## Test environment

NGINX Plus: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.4-gke.1043004
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 997.807µs
Latencies     [min, mean, 50, 90, 95, 99, max]  711.69µs, 957.047µs, 931.487µs, 1.041ms, 1.089ms, 1.247ms, 18.096ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 899.626µs
Latencies     [min, mean, 50, 90, 95, 99, max]  726.964µs, 985.028µs, 967.197µs, 1.079ms, 1.118ms, 1.252ms, 19.94ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 917.947µs
Latencies     [min, mean, 50, 90, 95, 99, max]  721.266µs, 993.472µs, 973.023µs, 1.101ms, 1.164ms, 1.321ms, 20.914ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.97
Duration      [total, attack, wait]             30s, 29.999s, 1.08ms
Latencies     [min, mean, 50, 90, 95, 99, max]  708.277µs, 986.926µs, 970.493µs, 1.103ms, 1.15ms, 1.274ms, 19.021ms
Bytes In      [total, mean]                     4739842, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 1.022ms
Latencies     [min, mean, 50, 90, 95, 99, max]  724.839µs, 981.279µs, 965.573µs, 1.107ms, 1.169ms, 1.299ms, 6.315ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
# Results

## Test environment

NGINX Plus: true

GKE Cluster:

- Node count: 12
- k8s version: v1.29.5-gke.1091002
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 813.859µs
Latencies     [min, mean, 50, 90, 95, 99, max]  562.969µs, 841.698µs, 802.462µs, 927.612µs, 978.803µs, 1.203ms, 14.688ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 780.7µs
Latencies     [min, mean, 50, 90, 95, 99, max]  597.539µs, 867.995µs, 841.23µs, 981.412µs, 1.054ms, 1.268ms, 25.427ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.06, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 876.487µs
Latencies     [min, mean, 50, 90, 95, 99, max]  595.128µs, 867.285µs, 846.087µs, 990.033µs, 1.047ms, 1.202ms, 11.997ms
Bytes In      [total, mean]                     5010000, 167.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 840.99µs
Latencies     [min, mean, 50, 90, 95, 99, max]  594.554µs, 862.747µs, 846.075µs, 971.413µs, 1.025ms, 1.204ms, 10.757ms
Bytes In      [total, mean]                     4680000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 826.77µs
Latencies     [min, mean, 50, 90, 95, 99, max]  616.343µs, 872.167µs, 853.157µs, 969.594µs, 1.016ms, 1.155ms, 24.084ms
Bytes In      [total, mean]                     4680000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
