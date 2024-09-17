# Results

## Test environment

NGINX Plus: false

NGINX Gateway Fabric:

- Commit: bf8ea47203eb4695af0d359243c73de2d1badbbf
- Date: 2024-09-13T20:33:11Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1639000
- vCPUs per node: 16
- RAM per node: 65853968Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 674.086µs
Latencies     [min, mean, 50, 90, 95, 99, max]  504.902µs, 690.958µs, 671.389µs, 776.467µs, 814.84µs, 920.64µs, 11.923ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 727.772µs
Latencies     [min, mean, 50, 90, 95, 99, max]  544.595µs, 724.644µs, 711.056µs, 815.628µs, 855.662µs, 971.898µs, 12.671ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 680.019µs
Latencies     [min, mean, 50, 90, 95, 99, max]  549.132µs, 745.892µs, 717.999µs, 829.205µs, 873.444µs, 1.024ms, 16.671ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 641.483µs
Latencies     [min, mean, 50, 90, 95, 99, max]  537.369µs, 718.158µs, 698.151µs, 811.422µs, 855.477µs, 967.85µs, 19.633ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 798.923µs
Latencies     [min, mean, 50, 90, 95, 99, max]  536.667µs, 719.1µs, 707.001µs, 811.278µs, 849.968µs, 956.036µs, 11.637ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
