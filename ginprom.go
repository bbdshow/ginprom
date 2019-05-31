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
	defaultSystemNS ="gin_prom"
	defaultSubSystem = "test"
)


type GinPrometheus struct {
	MetricsPath string
	engine *gin.Engine

	reqCnt *prometheus.CounterVec
	reqDur, reqSz, resSz prometheus.Summary


	SystemNS string
	Subsystem string

	// 固定路径
	fixedPath sync.Map

	// 带参数路径
	paramsPath sync.Map
}

type Config struct {
	SystemNS string
	Subsystem string
	MetricsPath string

	FixedPath []string
	ParamsPath map[string]int
}

func New(e *gin.Engine, cfg *Config) *GinPrometheus{
	if e == nil {
		panic("gin engine nil")
	}

	p := GinPrometheus{
		MetricsPath:defaultMetricsPath,
		SystemNS:defaultSystemNS,
		Subsystem:defaultSubSystem,
		fixedPath: sync.Map{},
		paramsPath: sync.Map{},
	}

	if cfg.MetricsPath != "" {
		p.MetricsPath = cfg.MetricsPath
	}
	if cfg.SystemNS != "" {
		p.SystemNS = cfg.SystemNS
	}

	if cfg.Subsystem != "" {
		p.Subsystem = cfg.Subsystem
	}

	for _, v := range cfg.FixedPath {
			p.SetFixedPath(v)
	}

	for k, v := range cfg.ParamsPath {
		p.SetParamsPath(k,v )
	}

	p.register()

	return &p
}

func (gp *GinPrometheus) Use(e *gin.Engine){
	e.GET(gp.MetricsPath, prometheusHandler())
	gp.engine = e
}

func (gp *GinPrometheus) SetFixedPath(path string){
	_, err := url.Parse(path)
	if err == nil {
		gp.fixedPath.Store(path, struct {}{})
	}
}

func (gp *GinPrometheus)DelFixedPath(path string){
	gp.fixedPath.Delete(path)
}

// index 占位符相对路由位置 1是起始位置  / 分割数组
func (gp *GinPrometheus) SetParamsPath(path string, index int){
	_, err := url.Parse(path)
	if err == nil {
		gp.paramsPath.Store(path, index)
	}
}

func (gp *GinPrometheus) DelParamsPath(path string) {
	gp.paramsPath.Delete(path)
}

func (gp *GinPrometheus) HandlerFunc() gin.HandlerFunc{
	return func(c *gin.Context) {
		path := c.Request.URL.String()
		path, ok := gp.HitPath(path)
		if !ok {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := gp.reqSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Millisecond)
		resSz := float64(c.Writer.Size())

		gp.reqDur.Observe(elapsed)

		gp.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, path).Inc()
		gp.reqSz.Observe(float64(reqSz))
		gp.resSz.Observe(resSz)
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


func (gp *GinPrometheus) reqSize(r *http.Request) int{
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