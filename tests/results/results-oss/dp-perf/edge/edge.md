# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.4-gke.1043004
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-east1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 986.95µs
Latencies     [min, mean, 50, 90, 95, 99, max]  562.91µs, 940.576µs, 912.579µs, 1.141ms, 1.212ms, 1.375ms, 12.402ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.97
Duration      [total, attack, wait]             30s, 29.999s, 947.18µs
Latencies     [min, mean, 50, 90, 95, 99, max]  602.01µs, 984.263µs, 961.474µs, 1.162ms, 1.221ms, 1.367ms, 16.033ms
Bytes In      [total, mean]                     4799840, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         29999, 1000.03, 1000.00
Duration      [total, attack, wait]             29.999s, 29.998s, 875.67µs
Latencies     [min, mean, 50, 90, 95, 99, max]  596.74µs, 998.15µs, 983.222µs, 1.188ms, 1.249ms, 1.386ms, 11.616ms
Bytes In      [total, mean]                     5039832, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.97
Duration      [total, attack, wait]             30.001s, 30s, 1.139ms
Latencies     [min, mean, 50, 90, 95, 99, max]  611.07µs, 1.005ms, 983.916µs, 1.193ms, 1.257ms, 1.41ms, 19.537ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 967.36µs
Latencies     [min, mean, 50, 90, 95, 99, max]  595.57µs, 999.204µs, 982.831µs, 1.192ms, 1.255ms, 1.392ms, 14.948ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
# Results

## Test environment

NGINX Plus: false

GKE Cluster:

- Node count: 12
- k8s version: v1.29.5-gke.1091002
- vCPUs per node: 16
- RAM per node: 65855012Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 899.476µs
Latencies     [min, mean, 50, 90, 95, 99, max]  583.648µs, 822.733µs, 798.079µs, 921.749µs, 961.231µs, 1.077ms, 31.226ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 823.051µs
Latencies     [min, mean, 50, 90, 95, 99, max]  611.86µs, 845.841µs, 814.787µs, 919.281µs, 957.112µs, 1.074ms, 39.474ms
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         29999, 1000.00, 999.98
Duration      [total, attack, wait]             30s, 29.999s, 853.399µs
Latencies     [min, mean, 50, 90, 95, 99, max]  616.402µs, 829.004µs, 818.781µs, 918.739µs, 955.645µs, 1.06ms, 13.295ms
Bytes In      [total, mean]                     5069831, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:29999  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 742.84µs
Latencies     [min, mean, 50, 90, 95, 99, max]  611.445µs, 820.114µs, 804.022µs, 930.064µs, 972.12µs, 1.067ms, 10.66ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.99
Duration      [total, attack, wait]             30s, 30s, 716.852µs
Latencies     [min, mean, 50, 90, 95, 99, max]  591.199µs, 821.731µs, 803.487µs, 929.572µs, 980.906µs, 1.09ms, 23.518ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
