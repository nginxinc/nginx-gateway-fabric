# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 747a8c8cb51d72104b88598068f4b7de330c3981
- Date: 2024-09-03T14:51:18Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.7-gke.1104000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 635.089µs
Latencies     [min, mean, 50, 90, 95, 99, max]  476.925µs, 667.998µs, 656.292µs, 746.828µs, 778.924µs, 860.107µs, 12.062ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 628.171µs
Latencies     [min, mean, 50, 90, 95, 99, max]  529.939µs, 701.597µs, 690.442µs, 786.674µs, 822.114µs, 910.842µs, 9.89ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.06, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 658.589µs
Latencies     [min, mean, 50, 90, 95, 99, max]  527.154µs, 713.359µs, 699.601µs, 804.852µs, 843.397µs, 927.976µs, 9.38ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 685.054µs
Latencies     [min, mean, 50, 90, 95, 99, max]  530.743µs, 703.953µs, 688.22µs, 791.622µs, 832.34µs, 934.799µs, 10.54ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 630.725µs
Latencies     [min, mean, 50, 90, 95, 99, max]  521.167µs, 715.243µs, 698.731µs, 798.688µs, 835.025µs, 934.171µs, 18.629ms
Bytes In      [total, mean]                     4739842, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```
