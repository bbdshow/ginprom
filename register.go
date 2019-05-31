package ginprom

import "github.com/prometheus/client_golang/prometheus"


func (gp *GinPrometheus)AddRegister(cs ...prometheus.Collector) {
		prometheus.MustRegister(cs...)
}

func (gp *GinPrometheus) register(){
	gp.reqCnt  = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: gp.SystemNS,
			Subsystem: gp.Subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "handler", "host", "path"},
	)

	gp.reqDur = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:gp.SystemNS,
			Subsystem:gp.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
		}, )

	gp.reqSz = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:gp.SystemNS,
			Subsystem:gp.Subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
	)

	gp.resSz = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:gp.SystemNS,
			Subsystem:gp.Subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
	)

	prometheus.MustRegister(gp.reqCnt, gp.reqDur,gp.reqSz, gp.resSz)

}