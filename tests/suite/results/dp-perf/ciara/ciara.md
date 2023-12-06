# Results

## Test environment

GKE Cluster:

- Node count: 3
- k8s version: v1.27.3-gke.100
- vCPUs per node: 2
- RAM per node: 4022960Ki
- Max pods per node: 110
- Zone: europe-west2-b
- Instance Type: e2-medium

## Test1: Running latte path based routing

```console
Running 30s test @ http://cafe.example.com/latte
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.13ms    6.01ms  89.25ms   91.31%
    Req/Sec     0.98k   227.22     1.73k    75.50%
  Latency Distribution
     50%    4.52ms
     75%    6.66ms
     90%   11.17ms
     99%   33.18ms
  58349 requests in 30.01s, 20.81MB read
Requests/sec:   1944.02
Transfer/sec:    710.01KB
```

## Test2: Running coffee header based routing

```console
Running 30s test @ http://cafe.example.com/coffee
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.67ms    6.76ms 166.23ms   92.85%
    Req/Sec     0.86k   208.16     1.49k    72.50%
  Latency Distribution
     50%    5.08ms
     75%    7.72ms
     90%   11.61ms
     99%   32.73ms
  51667 requests in 30.02s, 18.48MB read
Requests/sec:   1721.04
Transfer/sec:    630.26KB
```

## Test3: Running coffee query based routing

```console
Running 30s test @ http://cafe.example.com/coffee?TEST=v2
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     7.55ms   10.38ms 174.94ms   94.17%
    Req/Sec     0.87k   170.21     1.42k    75.83%
  Latency Distribution
     50%    4.86ms
     75%    7.61ms
     90%   12.89ms
     99%   59.25ms
  51758 requests in 30.02s, 18.90MB read
Requests/sec:   1724.12
Transfer/sec:    644.85KB
```

## Test4: Running tea GET method based routing

```console
Running 30s test @ http://cafe.example.com/tea
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.99ms    9.37ms 231.45ms   94.96%
    Req/Sec     0.88k   200.12     1.44k    71.83%
  Latency Distribution
     50%    4.97ms
     75%    7.52ms
     90%   11.67ms
     99%   42.94ms
  52302 requests in 30.01s, 18.55MB read
Requests/sec:   1742.65
Transfer/sec:    633.07KB
```

## Test5: Running tea POST method based routing

```console
Running 30s test @ http://cafe.example.com/tea
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    30.96ms   41.49ms 225.04ms   80.96%
    Req/Sec   441.48    318.35     1.40k    62.48%
  Latency Distribution
     50%    7.10ms
     75%   50.03ms
     90%  106.44ms
     99%  141.63ms
  25998 requests in 30.03s, 9.22MB read
Requests/sec:    865.61
Transfer/sec:    314.46KB
```
