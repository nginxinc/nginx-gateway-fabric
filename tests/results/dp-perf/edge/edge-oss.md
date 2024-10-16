# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: 3a08fdafadfe0fb4a9c25679da1a1fcd6b181474
- Date: 2024-10-15T13:45:52Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1014001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 724.131µs
Latencies     [min, mean, 50, 90, 95, 99, max]  506.273µs, 701.347µs, 679.076µs, 779.282µs, 819.797µs, 976.515µs, 12.161ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 692.989µs
Latencies     [min, mean, 50, 90, 95, 99, max]  532.886µs, 721.993µs, 706.02µs, 812.366µs, 860.289µs, 1.018ms, 4.297ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 713.827µs
Latencies     [min, mean, 50, 90, 95, 99, max]  538.212µs, 722.813µs, 705.872µs, 813.629µs, 859.743µs, 1.009ms, 7.374ms
Bytes In      [total, mean]                     5100000, 170.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30s, 30s, 690.398µs
Latencies     [min, mean, 50, 90, 95, 99, max]  530.942µs, 709.483µs, 690.422µs, 800.878µs, 843.812µs, 975.935µs, 9.135ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 773.112µs
Latencies     [min, mean, 50, 90, 95, 99, max]  512.273µs, 709.133µs, 692.254µs, 793.501µs, 832.049µs, 968.373µs, 11.168ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
