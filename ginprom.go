package ginprom

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defMetricsPath = "/metrics"
)

type GinPrometheus struct {
	MetricsPath string

	// 存放 Metrics定义 key=metric.ID value=Metric
	metricsMap sync.Map

	staticPathMap  sync.Map
	dynamicPathMap sync.Map
}

type Config struct {
	// metrics 信息获取路由， 默认 /metrics
	MetricsPath string

	// 静态路径
	StaticPath []string

	// key=动态路径 value=:query所在索引位置 例如 /v1/:query 此路由则设置成  {"/v1/:query": 2}
	DynamicPath map[string]int
}

func NewGinProm(engine *gin.Engine, cfg *Config) (*GinPrometheus, error) {
	if engine == nil || cfg == nil {
		panic("gin engine and config required")
	}

	gp := GinPrometheus{
		MetricsPath:    defMetricsPath,
		metricsMap:     sync.Map{},
		staticPathMap:  sync.Map{},
		dynamicPathMap: sync.Map{},
	}

	if cfg.MetricsPath != "" {
		gp.MetricsPath = cfg.MetricsPath
	}
	var err error
	for _, v := range cfg.StaticPath {
		err = gp.SetStaticPath(v)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range cfg.DynamicPath {
		err = gp.SetDynamicPath(k, v)
		if err != nil {
			return nil, err
		}
	}

	return &gp, nil
}

// UsePrometheusHandler 注册获取 metrics 信息的路由
func UsePrometheusHandler(e *gin.Engine, metricsPath string) {
	e.GET(metricsPath, prometheusHandler())
}

func prometheusHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func (gp *GinPrometheus) SetStaticPath(path string) error {
	u, err := url.Parse(path)
	if err != nil {
		return err
	}
	gp.staticPathMap.Store(u.String(), struct{}{})

	return nil
}

func (gp *GinPrometheus) DelStaticPath(path string) {
	gp.staticPathMap.Delete(path)
}

// index 占位符相对路由位置 1是起始位置  / 分割数组
func (gp *GinPrometheus) SetDynamicPath(path string, index int) error {
	u, err := url.Parse(path)
	if err != nil {
		return err
	}
	gp.dynamicPathMap.Store(u.String(), index)

	return nil
}

func (gp *GinPrometheus) DelDynamicPath(path string) {
	gp.dynamicPathMap.Delete(path)
}

func (gp *GinPrometheus) GetMetrics(id string) (Metric, bool) {
	v, ok := gp.metricsMap.Load(id)
	if !ok {
		return Metric{}, ok

	}
	return v.(Metric), ok
}

func (gp *GinPrometheus) SetMetrics(m Metric) {
	gp.metricsMap.Store(m.ID, m)
}

// DefaultMetricsMid 必须要使用DefaultRegister注册后才能使用
func DefaultMetricsMid(gp *GinPrometheus) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		path, ok := gp.HitPath(path)
		if !ok {
			c.Next()
			return
		}

		start := time.Now()
		requestSize := CalcReqSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		responseSize := float64(c.Writer.Size())

		reqCnt, ok := gp.GetMetrics(IdReqCnt)
		if ok {
			reqCnt.Collector.(*prometheus.CounterVec).WithLabelValues(status, c.Request.Method, c.Request.Host, path).Inc()
		}

		reqSz, ok := gp.GetMetrics(IdReqSize)
		if ok {
			reqSz.Collector.(prometheus.Summary).Observe(float64(requestSize))
		}

		resSz, ok := gp.GetMetrics(IdResSize)
		if ok {
			resSz.Collector.(prometheus.Summary).Observe(float64(responseSize))
		}

		reqDur, ok := gp.GetMetrics(IdReqDur)
		if ok {
			reqDur.Collector.(prometheus.Summary).Observe(float64(elapsed))
		}
	}
}

func (gp *GinPrometheus) HitPath(path string) (str string, ok bool) {
	_, ok = gp.staticPathMap.Load(path)
	if ok {
		return path, ok
	}

	exists := false

	// 动态路径解析
	gp.dynamicPathMap.Range(func(k, v interface{}) bool {
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

func CalcReqSize(r *http.Request) int {
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
