# Results

## Test environment

GKE Cluster:

- Node count: 3
- k8s version: v1.27.5-gke.200
- vCPUs per node: 2
- RAM per node: 4022908Ki
- Max pods per node: 110
- Zone: europe-west2-b
- Instance Type: e2-medium

## Test1: Running latte path based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 1000.00
Duration      [total, attack, wait]             30s, 29.999s, 691.365µs
Latencies     [min, mean, 50, 90, 95, 99, max]  422.618µs, 4.133ms, 857.936µs, 3.183ms, 17.814ms, 94.115ms, 191.517ms
Bytes In      [total, mean]                     4800000, 160.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000
Error Set:
```

## Test2: Running coffee header based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 763.01
Duration      [total, attack, wait]             39.318s, 29.999s, 9.319s
Latencies     [min, mean, 50, 90, 95, 99, max]  500.704µs, 628.036ms, 1.086ms, 3.512s, 4.268s, 8.068s, 13.416s
Bytes In      [total, mean]                     4830000, 161.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000
Error Set:
```

## Test3: Running coffee query based routing

```text
Requests      [total, rate, throughput]         30000, 1000.01, 999.98
Duration      [total, attack, wait]             30.001s, 30s, 941.332µs
Latencies     [min, mean, 50, 90, 95, 99, max]  471.118µs, 5.118ms, 936.844µs, 6.1ms, 21.269ms, 103.533ms, 207.304ms
Bytes In      [total, mean]                     5070000, 169.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000
Error Set:
```

## Test4: Running tea GET method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.02, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 964.988µs
Latencies     [min, mean, 50, 90, 95, 99, max]  500.904µs, 7.32ms, 937.131µs, 6.46ms, 38.071ms, 132.996ms, 376.081ms
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000
Error Set:
```

## Test5: Running tea POST method based routing

```text
Requests      [total, rate, throughput]         30000, 1000.03, 999.99
Duration      [total, attack, wait]             30s, 29.999s, 1.102ms
Latencies     [min, mean, 50, 90, 95, 99, max]  468.277µs, 100.905ms, 1.133ms, 350.181ms, 794.519ms, 1.592s, 1.697s
Bytes In      [total, mean]                     4740000, 158.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000
Error Set:
```
