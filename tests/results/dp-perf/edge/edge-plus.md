# Results

## Test environment

NGINX Plus: true

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
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 732.867µs
Latencies     [min, mean, 50, 90, 95, 99, max]  547.322µs, 785.057µs, 744.919µs, 842.69µs, 882.368µs, 1.089ms, 18.98ms
Bytes In      [total, mean]                     4770000, 159.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.01
Duration      [total, attack, wait]             30s, 29.999s, 790.67µs
Latencies     [min, mean, 50, 90, 95, 99, max]  598.016µs, 811.696µs, 801.075µs, 902.134µs, 941.046µs, 1.058ms, 9.661ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.06, 1000.02
Duration      [total, attack, wait]             29.999s, 29.998s, 1.003ms
Latencies     [min, mean, 50, 90, 95, 99, max]  621.299µs, 838.718µs, 822.301µs, 926.487µs, 966.776µs, 1.097ms, 12.543ms
Bytes In      [total, mean]                     5040000, 168.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 873.228µs
Latencies     [min, mean, 50, 90, 95, 99, max]  621.329µs, 821.996µs, 807.425µs, 906.307µs, 945.368µs, 1.067ms, 18.616ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.00, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 760.981µs
Latencies     [min, mean, 50, 90, 95, 99, max]  580.722µs, 808.039µs, 797.683µs, 891.712µs, 926.73µs, 1.038ms, 4.494ms
Bytes In      [total, mean]                     4710000, 157.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000  
Error Set:
```
