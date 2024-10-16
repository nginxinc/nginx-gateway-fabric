# Results

## Test environment

NGINX Plus: true

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
Duration      [total, attack, wait]             30s, 29.999s, 621.777µs
Latencies     [min, mean, 50, 90, 95, 99, max]  493.496µs, 682.208µs, 666.971µs, 771.821µs, 812.592µs, 935.803µs, 9.613ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 696.203µs
Latencies     [min, mean, 50, 90, 95, 99, max]  529.204µs, 709.079µs, 694.929µs, 791.712µs, 834.953µs, 960.962µs, 9.219ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 659.516µs
Latencies     [min, mean, 50, 90, 95, 99, max]  542.362µs, 719.357µs, 701.71µs, 805.054µs, 848.72µs, 980.33µs, 11.954ms
Bytes In      [total, mean]                     5069831, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 677.532µs
Latencies     [min, mean, 50, 90, 95, 99, max]  509.01µs, 703.22µs, 689.619µs, 790.875µs, 832.349µs, 970.323µs, 7.744ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 687.367µs
Latencies     [min, mean, 50, 90, 95, 99, max]  520.251µs, 696.116µs, 682.682µs, 780.437µs, 820.818µs, 945.164µs, 8.67ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
