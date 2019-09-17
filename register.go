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

	IdReqCnt     = "reqCnt"
	IdReqDur     = "reqDur"
	IdResSize    = "resSize"
	IdReqSize    = "reqSize"
	IdReqElapsed = "reqElapsed"
)

type Metric struct {
	// Collector 根据Type通过 NewCollector 生成
	Collector   prometheus.Collector
	ID          string // metric 标志
	Name        string
	Description string
	Buckets     []float64
	Type        string
	Args        []string
}

var (
	ReqCnt = Metric{
		ID:          IdReqCnt,
		Name:        "requests_total",
		Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		Type:        CounterVecType,
		Args:        []string{"code", "method", "host", "path"}}

	ReqDur = Metric{
		ID:          IdReqDur,
		Name:        "request_duration_seconds",
		Description: "The HTTP request latencies in seconds.",
		Type:        SummaryType}

	ResSize = Metric{
		ID:          IdResSize,
		Name:        "response_size_bytes",
		Description: "The HTTP response sizes in bytes.",
		Type:        SummaryType}

	ReqSize = Metric{
		ID:          IdReqSize,
		Name:        "request_size_bytes",
		Description: "The HTTP request sizes in bytes.",
		Type:        SummaryType}

	ReqElapsed = Metric{
		ID:          IdReqElapsed,
		Name:        "http_request_elapsed_second",
		Description: "http api request elapsed",
		Type:        HistogramVecType,
		Args:        []string{"method", "path"},
	}
)

var defaultMetrics = []Metric{ReqCnt, ReqDur, ResSize, ReqSize, ReqElapsed}

// DefaultRegister subsystem 子系统，如果不存在，可以不填写(单一类型服务可以通过 jobname 区分开)
func (gp *GinPrometheus) DefaultMetricsRegister(subsystem string) {
	for i := range defaultMetrics {

		defaultMetrics[i].NewCollector(subsystem)

		gp.SetMetrics(defaultMetrics[i])

		prometheus.MustRegister(defaultMetrics[i].Collector)
	}
}

// AddMetrics 可以在默认的基础上增加其他的metrics
func (gp *GinPrometheus) AddMetrics(m Metric) error {
	_, ok := gp.GetMetrics(m.ID)
	if ok {
		return fmt.Errorf("%s exists", m.ID)
	}

	gp.SetMetrics(m)

	return prometheus.Register(m.Collector)
}

// NewCollector, Metric 必须通过次方法才能生效
func (m *Metric) NewCollector(subsystem string) {
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
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			m.Args,
		)
	case SummaryType:
		m.Collector = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
		)
	}

}
