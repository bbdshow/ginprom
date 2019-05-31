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

type Metric struct {
	Type            string
	Name            string
	Description     string
	Args            []string
}

func(gp *GinPrometheus) NewMetric(m *Metric) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Namespace: gp.SystemNS,
				Subsystem: gp.Subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}