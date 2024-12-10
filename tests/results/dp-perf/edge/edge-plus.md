# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 929413c15af7bee3adb32e103c9d1513a693da16
- Date: 2024-11-28T12:52:45Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.5-gke.1443001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 724.601µs
Latencies     [min, mean, 50, 90, 95, 99, max]  529.511µs, 729.144µs, 709.186µs, 818.381µs, 860.485µs, 980.243µs, 19.544ms
Bytes In      [total, mean]                     4859838, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 817.269µs
Latencies     [min, mean, 50, 90, 95, 99, max]  532.895µs, 754.301µs, 737.323µs, 855.307µs, 899.065µs, 1.022ms, 12.677ms
Bytes In      [total, mean]                     4890000, 163.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.05, 1000.03
Duration      [total, attack, wait]             29.999s, 29.998s, 723.164µs
Latencies     [min, mean, 50, 90, 95, 99, max]  545.543µs, 753.599µs, 739.694µs, 851.051µs, 895.471µs, 1.027ms, 14.092ms
Bytes In      [total, mean]                     5130000, 171.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 848.673µs
Latencies     [min, mean, 50, 90, 95, 99, max]  547.341µs, 762.917µs, 746.645µs, 858.635µs, 901.633µs, 1.023ms, 15.807ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 714.679µs
Latencies     [min, mean, 50, 90, 95, 99, max]  530.13µs, 757.22µs, 741.791µs, 858.692µs, 901.679µs, 1.026ms, 12.263ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
