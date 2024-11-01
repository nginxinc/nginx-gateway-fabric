# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: fed4239ecb35f937b66bba7bd68d6894ca0762b3
- Date: 2024-11-01T00:13:12Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1355000
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         29999, 1000.01, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 698.936µs
Latencies     [min, mean, 50, 90, 95, 99, max]  513.083µs, 738.127µs, 713.194µs, 822.227µs, 869.319µs, 991.029µs, 21.514ms
Bytes In      [total, mean]                     4799840, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 737.126µs
Latencies     [min, mean, 50, 90, 95, 99, max]  565.872µs, 760.88µs, 746.011µs, 862.068µs, 909.532µs, 1.038ms, 6.617ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 678.831µs
Latencies     [min, mean, 50, 90, 95, 99, max]  561.826µs, 780.678µs, 761.379µs, 891.349µs, 939.072µs, 1.059ms, 11.658ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 655.792µs
Latencies     [min, mean, 50, 90, 95, 99, max]  549.409µs, 770.46µs, 751.463µs, 871.937µs, 915.03µs, 1.042ms, 19.785ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 766.173µs
Latencies     [min, mean, 50, 90, 95, 99, max]  546.875µs, 771.691µs, 752.254µs, 871.276µs, 918.398µs, 1.04ms, 13.175ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
