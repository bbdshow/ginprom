package ginprom

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	CounterType      = "counter"
	CounterVecType   = "counter_vec"
	GaugeType        = "gauge"
	GaugeVecType     = "gauge_vec"
	HistogramType    = "histogram"
	HistogramVecType = "histogram_vec"
	SummaryType      = "summary"
	SummaryVecType   = "summary_vec"
)

type Metric struct {
	Collector   prometheus.Collector
	ID          string // metric 标志
	Name        string
	Description string
	Buckets     []float64
	Type        string
	Args        []string
}

var (
	reqCnt = Metric{
		ID:          "reqCnt",
		Name:        "requests_total",
		Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		Type:        CounterVecType,
		Args:        []string{"code", "method", "handler", "host", "path"}}

	reqDur = Metric{
		ID:          "reqDur",
		Name:        "request_duration_seconds",
		Description: "The HTTP request latencies in seconds.",
		Type:        SummaryType}

	resSz = Metric{
		ID:          "resSz",
		Name:        "response_size_bytes",
		Description: "The HTTP response sizes in bytes.",
		Type:        SummaryType}

	reqSz = Metric{
		ID:          "reqSz",
		Name:        "request_size_bytes",
		Description: "The HTTP request sizes in bytes.",
		Type:        SummaryType}
)

var defaultMetrics = []Metric{reqCnt, reqDur, resSz, reqSz}

func (gp *GinPrometheus) DefaultRegister(subsystem string) {
	for i := range defaultMetrics {
		NewMetric(&defaultMetrics[i], subsystem)

		gp.MetricsMap.Store(defaultMetrics[i].ID, defaultMetrics[i])

		prometheus.MustRegister(defaultMetrics[i].Collector)

	}
}

// 不是默认的ID才注册
func (gp *GinPrometheus) AddMetrics(m *Metric, subsystem string) error {
	_, ok := gp.MetricsMap.Load(m.ID)
	if ok {
		return fmt.Errorf("%s exists", m.ID)
	}
	NewMetric(m, subsystem)
	gp.MetricsMap.Store(m.ID, *m)

	return prometheus.Register(m.Collector)
}

func NewMetric(m *Metric, subsystem string) {
	switch m.Type {
	case CounterVecType:
		m.Collector = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case CounterType:
		m.Collector = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case GaugeVecType:
		m.Collector = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case GaugeType:
		m.Collector = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case HistogramVecType:
		m.Collector = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   m.Buckets,
			},
			m.Args,
		)
	case HistogramType:
		m.Collector = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Buckets:   m.Buckets,
			},
		)
	case SummaryVecType:
		m.Collector = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case SummaryType:
		m.Collector = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}

}
