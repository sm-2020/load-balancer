# load-balancer
> A basic load balancer implementation in golang

Load Balancers have different strategies for distributing the load across a set of backends.
- <b>Round Robin</b>  distribute load equally, assumes all backends have the same processing power
- <b>Weighted Round Robin</b> - Additional weights can be given considering the backend’s processing power
- <b>Least Connections</b> - Load is distributed to the servers with least active connections
