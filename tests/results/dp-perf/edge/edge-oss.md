# Results

## Test environment

NGINX Plus: false

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
Requests      [total, rate, throughput]         30000, 1000.02, 998.33
Duration      [total, attack, wait]             30s, 29.999s, 587.493µs
Latencies     [min, mean, 50, 90, 95, 99, max]  300.306µs, 659.433µs, 640.281µs, 723.707µs, 754.811µs, 848.134µs, 18.652ms
Bytes In      [total, mean]                     4799500, 159.98
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           99.83%
Status Codes  [code:count]                      200:29950  502:50  
Error Set:
502 Bad Gateway
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 649.364µs
Latencies     [min, mean, 50, 90, 95, 99, max]  500.587µs, 690.268µs, 664.953µs, 756.292µs, 794.122µs, 881.61µs, 22.865ms
Bytes In      [total, mean]                     4829839, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 628.951µs
Latencies     [min, mean, 50, 90, 95, 99, max]  528.846µs, 682.665µs, 668.807µs, 755.238µs, 788.506µs, 867.725µs, 12.241ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 693.882µs
Latencies     [min, mean, 50, 90, 95, 99, max]  506.955µs, 686.408µs, 670.886µs, 759.153µs, 792.745µs, 886.171µs, 17.088ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 540.887µs
Latencies     [min, mean, 50, 90, 95, 99, max]  505.067µs, 688.241µs, 670.462µs, 765.138µs, 803.564µs, 904.637µs, 20.383ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
