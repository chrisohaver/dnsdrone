package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	qps                int
	queries            []dns.Question
	timeout            time.Duration
	verbose            bool
	localResolver      bool
	namelist, promaddr string
	delay              time.Duration

	RequestCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "dnsdrone",
		Name:      "request_count_total",
		Help:      "Counter of DNS requests sent",
	})

	ResponseCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "dnsdrone",
		Name:      "response_count_total",
		Help:      "Counter of DNS responses",
	}, []string{"rcode"})

	ResponseLostCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "dnsdrone",
		Name:      "response_lost_count_total",
		Help:      "Counter of DNS responses lost",
	})

	RequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "dnsdrone",
		Name:      "request_duration_seconds",
		Buckets:   prometheus.ExponentialBuckets(0.00025, 2, 16), // from 0.25ms to 8 seconds
		Help:      "Histogram of the time (in seconds) each request took",
	})
)

func main() {

	flag.IntVar(&qps, "qps", 1, "DNS queries per second")
	flag.StringVar(&promaddr, "prom", ":9696", "Prometheus endpoint")
	flag.StringVar(&namelist, "names", "", "Comma separated list of hostnames")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "Timeout for DNS queries (for non-local resolver)")
	flag.BoolVar(&verbose, "verbose", false, "Verbose log output")
	flag.BoolVar(&localResolver, "local-resolver", true, "Use local resolver")
	flag.DurationVar(&delay, "delay", 0, "Time to wait before sending queries")

	flag.Parse()

	for _, name := range strings.Split(namelist, ",") {
		if name == "" {
			continue
		}
		// names from command line flag default to type A
		queries = append(queries, dns.Question{Name: name, Qtype: dns.TypeA})
	}

	if len(queries) == 0 {
		log.Fatal("No query names found")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(1 * time.Second / time.Duration(qps))
	defer ticker.Stop()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(promaddr, nil)

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		log.Fatal(err)
	}
	address := config.Servers[0] + ":" + config.Port
	c := new(dns.Client)

	if !localResolver {
		c.Timeout = timeout
	}

	i := 0
	end := len(queries) - 1

	time.Sleep(delay)

	log.Printf("Sending %v queries per second to %v", qps, address)
	for {
		select {
		case <-ticker.C:
			go func() {
				debugf("Sending query %v type %v", queries[i].Name, dns.TypeToString[queries[i].Qtype])
				RequestCount.Inc()
				if localResolver {
					start := time.Now()
					_, err := net.LookupIP(queries[i].Name)
					rtt := time.Since(start)
					RequestDuration.Observe(rtt.Seconds())
					debugf("Received response in %v milliseconds", rtt.Milliseconds())
					if err == nil {
						ResponseCount.WithLabelValues("NOERROR").Inc()
						return
					}
					if dnserr, ok := err.(*net.DNSError); ok && dnserr.IsTimeout {
						ResponseLostCount.Inc()
						return
					}
					if dnserr, ok := err.(*net.DNSError); ok && dnserr.IsNotFound {
						ResponseCount.WithLabelValues("NXDOMAIN").Inc()
						return
					}
					ResponseCount.WithLabelValues("other").Inc()
					return
				}

				m := new(dns.Msg)
				m.SetQuestion(queries[i].Name, queries[i].Qtype)

				r, rtt, err := c.Exchange(m, address)
				if err != nil {
					debugf("Error sending query: %v", err)
					ResponseLostCount.Inc()
					return
				}
				debugf("Received response type %v in %v milliseconds", dns.RcodeToString[r.Rcode], rtt.Milliseconds())
				RequestDuration.Observe(rtt.Seconds())
				ResponseCount.WithLabelValues(dns.RcodeToString[r.Rcode]).Inc()

			}()
		case <-sig:
			log.Printf("Got signal, exiting")
			os.Exit(0)
		}

		if i >= end {
			i = 0
			continue
		}
		i++
	}
}

func debugf(fmt string, v ...interface{}) {
	if !verbose {
		return
	}
	log.Printf(fmt, v...)
}
