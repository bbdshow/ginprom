package ginprom

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	defaultMetricsPath = "/metrics"
)

type GinPrometheus struct {
	MetricsPath string
	engine      *gin.Engine

	MetricsMap sync.Map // map[string]Metric // key=metric.ID

	// 固定路径
	fixedPath sync.Map

	// 带参数路径
	paramsPath sync.Map
}

type Config struct {
	MetricsPath string

	FixedPath  []string
	ParamsPath map[string]int
}

func New(e *gin.Engine, cfg *Config) *GinPrometheus {
	if e == nil {
		panic("gin engine nil")
	}

	p := GinPrometheus{
		MetricsPath: defaultMetricsPath,
		MetricsMap:  sync.Map{},
		fixedPath:   sync.Map{},
		paramsPath:  sync.Map{},
	}

	if cfg.MetricsPath != "" {
		p.MetricsPath = cfg.MetricsPath
	}

	for _, v := range cfg.FixedPath {
		p.SetFixedPath(v)
	}

	for k, v := range cfg.ParamsPath {
		p.SetParamsPath(k, v)
	}

	return &p
}

func (gp *GinPrometheus) Use(e *gin.Engine) {
	e.GET(gp.MetricsPath, prometheusHandler())
	gp.engine = e
}

func (gp *GinPrometheus) SetFixedPath(path string) {
	_, err := url.Parse(path)
	if err == nil {
		gp.fixedPath.Store(path, struct{}{})
	}
}

func (gp *GinPrometheus) DelFixedPath(path string) {
	gp.fixedPath.Delete(path)
}

// index 占位符相对路由位置 1是起始位置  / 分割数组
func (gp *GinPrometheus) SetParamsPath(path string, index int) {
	_, err := url.Parse(path)
	if err == nil {
		gp.paramsPath.Store(path, index)
	}
}

func (gp *GinPrometheus) DelParamsPath(path string) {
	gp.paramsPath.Delete(path)
}

// DefaultHandlerFunc 必须要使用DefaultRegister注册后才能使用
func (gp *GinPrometheus) DefaultHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		path, ok := gp.HitPath(path)
		if !ok {
			c.Next()
			return
		}

		start := time.Now()
		requestSize := gp.ReqSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		responseSize := float64(c.Writer.Size())

		reqCnt, ok := gp.MetricsMap.Load("reqCnt")
		if ok {
			reqCnt.(Metric).Collector.(*prometheus.CounterVec).WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, path).Inc()
		}

		reqSz, ok := gp.MetricsMap.Load("reqSz")
		if ok {
			reqSz.(Metric).Collector.(prometheus.Summary).Observe(float64(requestSize))
		}

		resSz, ok := gp.MetricsMap.Load("resSz")
		if ok {
			resSz.(Metric).Collector.(prometheus.Summary).Observe(float64(responseSize))
		}

		reqDur, ok := gp.MetricsMap.Load("reqDur")
		if ok {
			reqDur.(Metric).Collector.(prometheus.Summary).Observe(float64(elapsed))
		}
	}
}

func (gp *GinPrometheus) HitPath(path string) (str string, ok bool) {
	_, ok = gp.fixedPath.Load(path)
	if ok {
		return path, ok
	}

	exists := false

	// 带参数的路径
	gp.paramsPath.Range(func(k, v interface{}) bool {
		index := v.(int)

		strs := strings.Split(path, "/")
		kstrs := strings.Split(k.(string), "/")

		if len(strs) > index {
			strs = append(strs[:index], strs[index+1:]...)
		}
		if len(kstrs) > index {
			kstrs = append(kstrs[:index], kstrs[index+1:]...)
		}

		if strings.Join(strs, "/") == strings.Join(kstrs, "/") {
			path = k.(string)
			exists = true
			return false
		}
		return true
	})

	return path, exists
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (gp *GinPrometheus) ReqSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)

	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}

	return s
}
