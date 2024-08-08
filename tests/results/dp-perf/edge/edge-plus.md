# Results

## Test environment

NGINX Plus: true

 NGINX Gateway Fabric:

- Commit: unknown
- Date: unknown
- Dirty: unknown

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1254000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 721.633µs
Latencies     [min, mean, 50, 90, 95, 99, max]  526.503µs, 690.211µs, 673.955µs, 769.117µs, 807.759µs, 919.561µs, 10.448ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 730.971µs
Latencies     [min, mean, 50, 90, 95, 99, max]  538.955µs, 714.812µs, 701.536µs, 799.398µs, 839.599µs, 949.93µs, 8.457ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 30s, 696.566µs
Latencies     [min, mean, 50, 90, 95, 99, max]  533.903µs, 729.146µs, 714.103µs, 817.188µs, 860.053µs, 970.953µs, 13.316ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 813.871µs
Latencies     [min, mean, 50, 90, 95, 99, max]  546.807µs, 722.997µs, 706.697µs, 810.134µs, 852.748µs, 965.725µs, 12.535ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 720.872µs
Latencies     [min, mean, 50, 90, 95, 99, max]  545.364µs, 734.795µs, 721.58µs, 828.75µs, 871.855µs, 990.283µs, 7.857ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
