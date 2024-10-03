# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: d7d6b0af0d56721b28aba24c1541d650ef6bc5a9
- Date: 2024-09-30T23:47:54Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1969001
- vCPUs per node: 16
- RAM per node: 65853972Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 613.524µs
Latencies     [min, mean, 50, 90, 95, 99, max]  493.667µs, 661.552µs, 643.908µs, 726.654µs, 763.9µs, 862.035µs, 16.392ms
Bytes In      [total, mean]                     4829839, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 688.713µs
Latencies     [min, mean, 50, 90, 95, 99, max]  522.959µs, 689.862µs, 672.734µs, 764.478µs, 797.131µs, 890.809µs, 12.189ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 670.956µs
Latencies     [min, mean, 50, 90, 95, 99, max]  531.378µs, 688.422µs, 674.945µs, 759.955µs, 793.679µs, 885.247µs, 14.82ms
Bytes In      [total, mean]                     5100000, 170.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 750.387µs
Latencies     [min, mean, 50, 90, 95, 99, max]  530.656µs, 682.96µs, 663.291µs, 755.22µs, 792.507µs, 882.755µs, 20.103ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 648.522µs
Latencies     [min, mean, 50, 90, 95, 99, max]  519.81µs, 675.008µs, 658.634µs, 742.332µs, 782.502µs, 869.715µs, 12.734ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
