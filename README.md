# load-balancer
> A basic load balancer implementation in golang

Load Balancers have different strategies for distributing the load across a set of backends.
- <b>Round Robin</b>  distribute load equally, assumes all backends have the same processing power
- <b>Weighted Round Robin</b> - Additional weights can be given considering the backend’s processing power
- <b>Least Connections</b> - Load is distributed to the servers with least active connections

This implementation uses RoundRobin algorithm to send requests into set of backends and support retries too.It also performs active cleaning and passive recovery for unhealthy backends.
Since its simple it assume if / is reachable for any host its available

### How to use
```
Usage:
  -backends string
        Load balanced backends, use commas to separate
  -port int
        Port to serve (default 3030)
```
