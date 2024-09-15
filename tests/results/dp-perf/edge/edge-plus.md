# Results

## Test environment

NGINX Plus: true

NGINX Gateway Fabric:

- Commit: bf8ea47203eb4695af0d359243c73de2d1badbbf
- Date: 2024-09-13T20:33:11Z
- Dirty: false

GKE Cluster:

- Node count: 12
- k8s version: v1.30.3-gke.1639000
- vCPUs per node: 16
- RAM per node: 65853960Ki
- Max pods per node: 110
- Zone: us-west1-b
- Instance Type: n2d-standard-16

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 659.753µs
Latencies     [min, mean, 50, 90, 95, 99, max]  516.947µs, 682.043µs, 667.613µs, 764.261µs, 807.989µs, 925.905µs, 10.402ms
Bytes In      [total, mean]                     4860000, 162.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.04, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 672.01µs
Latencies     [min, mean, 50, 90, 95, 99, max]  544.056µs, 708.524µs, 693.659µs, 798.224µs, 842.354µs, 962.69µs, 10.503ms
Bytes In      [total, mean]                     4890000, 163.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 688.255µs
Latencies     [min, mean, 50, 90, 95, 99, max]  548.294µs, 725.822µs, 704.967µs, 812.886µs, 862.452µs, 1.002ms, 16.667ms
Bytes In      [total, mean]                     5130000, 171.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 618.502µs
Latencies     [min, mean, 50, 90, 95, 99, max]  540.088µs, 700.199µs, 685.361µs, 780.736µs, 823.779µs, 952µs, 10.406ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 696.35µs
Latencies     [min, mean, 50, 90, 95, 99, max]  533.042µs, 705.177µs, 691.789µs, 786.608µs, 828.257µs, 945.722µs, 11.154ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
