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

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             30s, 29.999s, 600.749µs
Latencies     [min, mean, 50, 90, 95, 99, max]  492.509µs, 691.607µs, 658.092µs, 789.98µs, 846.577µs, 987.091µs, 18.354ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 664.964µs
Latencies     [min, mean, 50, 90, 95, 99, max]  515.752µs, 691.172µs, 665.785µs, 804.781µs, 869.797µs, 1.036ms, 10.611ms
Bytes In      [total, mean]                     4769841, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 639.03µs
Latencies     [min, mean, 50, 90, 95, 99, max]  527.807µs, 716.574µs, 691.757µs, 842.519µs, 916.915µs, 1.077ms, 6.682ms
Bytes In      [total, mean]                     5010000, 167.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 790.543µs
Latencies     [min, mean, 50, 90, 95, 99, max]  515.957µs, 681.264µs, 656.878µs, 788.001µs, 850.507µs, 1.011ms, 9.947ms
Bytes In      [total, mean]                     4680000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 653.438µs
Latencies     [min, mean, 50, 90, 95, 99, max]  509.058µs, 706.587µs, 682.67µs, 835.627µs, 909.532µs, 1.065ms, 11.152ms
Bytes In      [total, mean]                     4680000, 156.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
