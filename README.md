dnsdrone is a testing tool that creates an on going load of DNS queries on a server.

It produces metrics via Prometheus.

```
Usage of ./dnsdrone:
  -names string
        Comma separated list of hostnames
  -prom string
        prometheus endpoint (default ":9696")
  -qps int
        DNS queries per second (default 10)
  -timeout duration
        Timeout for DNS queries (default 1s)
  -verbose
        Verbose log output
```
