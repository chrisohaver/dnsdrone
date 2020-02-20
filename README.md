dnsdrone is a testing tool that creates an on going dns load on a server.

By default, it will generate a constant DNS query load to it's DNS server.
It produces metrics via Prometheus including:
 * dnsdrone_request_count
 * dnsdrone_request_duration
 * dnsdrone_response_count{rcode}
 * dnsdrone_response_lost_count