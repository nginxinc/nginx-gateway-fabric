# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: 9a85dbcc0797e31557a3731688795aa166ee0f96
- Date: 2024-08-13T21:12:05Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.6-gke.1326000
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 682.261µs
Latencies     [min, mean, 50, 90, 95, 99, max]  515.366µs, 729.019µs, 712.033µs, 843.484µs, 895.476µs, 1.019ms, 9.154ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 599.51µs
Latencies     [min, mean, 50, 90, 95, 99, max]  540.423µs, 740.62µs, 724.16µs, 857.72µs, 911.755µs, 1.042ms, 7.517ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.02
Duration      [total, attack, wait]             29.999s, 29.999s, 683.488µs
Latencies     [min, mean, 50, 90, 95, 99, max]  523.729µs, 756.497µs, 740.664µs, 870.743µs, 921.816µs, 1.062ms, 12.337ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 762.485µs
Latencies     [min, mean, 50, 90, 95, 99, max]  499.61µs, 744.201µs, 726.564µs, 862.287µs, 914.928µs, 1.054ms, 7.531ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 651.893µs
Latencies     [min, mean, 50, 90, 95, 99, max]  532.673µs, 751.908µs, 734.466µs, 872.179µs, 927.874µs, 1.068ms, 6.327ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
