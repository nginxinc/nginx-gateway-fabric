# Test Results Summary

## Version 1.0

### Scale Listeners

| Total Reloads | Total Reload Errors | Average Reload Time (ms) |
|---------------|---------------------|--------------------------|
| 107           | 0                   | 155.05374791154367       |

**NGINX Errors**: None.

**NGF Errors**: None.

**Pod Restarts**: None.

**CPU**: Steep linear increase as NGF processed all the Services. Dropped off during scaling of Listeners.
See [graph](/tests/scale/results/1.0/TestScale_Listeners/CPU.png).

**Memory**: Gradual increase in memory. Topped out at 40MiB.
See [graph](/tests/scale/results/1.0/TestScale_Listeners/Memory.png).

**Time To Ready**: Time to ready numbers consistently under 3s. 62nd Listener had longest TTR of 3.02s.
See [graph](/tests/scale/results/1.0/TestScale_Listeners/TTR.png).

### Scale HTTPS Listeners

| Total Reloads | Total Reload Errors | Average Reload Time (ms) |
|---------------|---------------------|--------------------------|
| 130           | 0                   | 151.49397590361446ms     |

**NGINX Errors**: None.

**NGF Errors**: None.

**Pod Restarts**: None.

**CPU**: Steep linear increase as NGF processed all the Services and Secrets. Dropped off during scaling of Listeners.
See [graph](/tests/scale/results/1.0/TestScale_HTTPSListeners/CPU.png).

**Memory**: Mostly linear increase. Topping out at right under 50MiB.
See [graph](/tests/scale/results/1.0/TestScale_HTTPSListeners/Memory.png).

**Time To Ready**: The time to ready numbers were pretty consistent (under 3 sec) except for one spike of 10s. I believe
this spike was client-side because the NGF logs indicated that the reload successfully happened under 3s.
See [graph](/tests/scale/results/1.0/TestScale_HTTPSListeners/TTR.png).

### Scale HTTPRoutes

| Total Reloads | Total Reload Errors | Average Reload Time (ms) |
|---------------|---------------------|--------------------------|
| 1001          | 0                   | 354.3878787878788ms      |

**NGINX Errors**: None.

**NGF Errors**: None.

**Pod Restarts**: None.

**CPU**: CPU mostly oscillated between .04 and .06. Several spikes over .06.
See [graph](/tests/scale/results/1.0/TestScale_HTTPRoutes/CPU.png).

**Memory**: Memory usage gradually increased from 25 - 150MiB over course of the test with some spikes reaching up to
200MiB. See [graph](/tests/scale/results/1.0/TestScale_HTTPRoutes/Memory.png).

**Time To Ready**: This time to ready graph is unique because there are three plotted lines:

- Blue Line: 2-second delay after adding a new HTTPRoute.
- Red Line: No delay after adding a new HTTPRoute.
- Green Line: 10-second delay after adding a new HTTPRoute

The Blue and Red lines are incomplete because the tests timed out. However, I think the implications are pretty clear.
The more time that passes between scaling events, the smaller the time to ready values are. This is because NGF
re-queues all the HTTPRoutes after updating their statuses. This is because the HTTPRoute has changed after we write its
status. This is compounded by the fact that NGF writes status for every HTTPRoute in the graph on every configuration
update. So if you add HTTPRoute 100, NGF will update the configuration with this new route and then update the status of
all 100 HTTPRoutes in the graph.

Related issues:

- https://github.com/nginxinc/nginx-gateway-fabric/issues/1013
- https://github.com/nginxinc/nginx-gateway-fabric/issues/825

See [graph](/tests/scale/results/1.0/TestScale_HTTPRoutes/TTR.png).

### Scale Upstream Servers

| Start Time (UNIX) | End Time (UNIX) | Duration (s) | Total Reloads | Total Reload Errors | Average Reload Time (ms) |
|-------------------|-----------------|--------------|---------------|---------------------|--------------------------|
| 1696535183        | 1696535311      | 128          | 83            | 0                   | 126.55555555555557       |

**NGINX Errors**: None.

**NGF Errors**: None.

**Pod Restarts**: None.

**CPU**: CPU steeply increases as NGF handles all the new Pods. Drops after they are processed.
See [graph](/tests/scale/results/1.0/TestScale_UpstreamServers/CPU.png).

**Memory**: Memory stays relatively flat and under 40MiB.
See [graph](/tests/scale/results/1.0/TestScale_UpstreamServers/Memory.png).

### Scale HTTP Matches

**Results for the first match**:

```text
Running 30s test @ http://cafe.example.com
2 threads and 10 connections
Thread Stats   Avg      Stdev     Max       +/- Stdev
Latency       47.64ms   13.87ms   217.49ms   97.52%
Req/Sec      107.84     17.47     151.00     79.76%
6410 requests in 30.09s, 0.95MB read
Requests/sec: 213.02
Transfer/sec: 32.24KB
```

**Results for the last match**:

```text
Running 30s test @ http://cafe.example.com
2 threads and 10 connections
Thread Stats   Avg       Stdev     Max      +/- Stdev
Latency       47.10ms    13.59ms  301.73ms   98.57%
Req/Sec       108.01     12.55    150.00     84.62%
6459 requests in 30.10s, 0.95MB read
Requests/sec:  214.61
Transfer/sec:  32.49KB
```

**Findings**:

- There's not a noticeable difference between the response times for the first match and last match. In
fact, the latency of the last match is slightly lower than the latency of the first match.
- If you add one more match to the [manifest](/tests/scale/manifests/scale-matches.yaml) nginx will fail to reload
  because the generate `http_matches` variable is too long.

Issue Filed: https://github.com/nginxinc/nginx-gateway-fabric/issues/1107
