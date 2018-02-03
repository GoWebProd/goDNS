package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	totalRequestsTcp = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "total",
		Help:      "total requests",

		ConstLabels: map[string]string{
			"type": "tcp",
		},
	}))

	totalRequestsUdp = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "total",
		Help:      "total requests",

		ConstLabels: map[string]string{
			"type": "udp",
		},
	}))

	totalRequestsFailed = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "failed",
		Help:      "failed requests",
	}))

	totalRequestsBlocked = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "blocked",
		Help:      "blocked requests",
	}))

	totalRequestsSuccess = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "success",
		Help:      "success requests",
	}))
)

func runPrometheus() {
	prometheus.MustRegister(totalRequestsTcp)
	prometheus.MustRegister(totalRequestsUdp)
	prometheus.MustRegister(totalRequestsFailed)
	prometheus.MustRegister(totalRequestsBlocked)
	prometheus.MustRegister(totalRequestsSuccess)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9970", nil))
}
