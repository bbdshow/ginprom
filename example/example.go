package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/huzhongqing/ginprom"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	engine := gin.New()

	cfg := ginprom.Config{
		StaticPath: []string{"/v1/hello"},
		DynamicPath: map[string]int{
			"/v1/hello/:girl":        3,
			"/v1/call/handsome/:boy": 4,
		},
	}

	prom, err := ginprom.NewGinProm(engine, &cfg)
	if err != nil {
		log.Println(err)
		return
	}

	prom.DefaultMetricsRegister("")

	ginprom.UsePrometheusHandler(engine, prom.MetricsPath)

	// 注册默认的 Metrics 中间件
	engine.Use(ginprom.DefaultMetricsMid(prom))

	// 自定义自己想要的结构
	iMetric := ginprom.Metric{
		ID:          "PV",
		Name:        "requests_pv",
		Description: "requests pv",
		Type:        "counter",
	}
	iMetric.NewCollector("")

	if err := prom.AddMetrics(iMetric); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	pvMid := func(v *ginprom.GinPrometheus) gin.HandlerFunc {
		return func(c *gin.Context) {
			pv, ok := v.GetMetrics("PV")
			if ok {
				pv.Collector.(prometheus.Counter).Inc()
			}
			c.Next()
		}
	}

	engine.Use(pvMid(prom))

	engine.GET("/v1/hello", func(c *gin.Context) {
		c.JSON(200, "hello ginprom!")
	})

	engine.GET("/v1/hello/:girl", func(c *gin.Context) {
		c.JSON(200, "hello "+c.Param("girl"))
	})

	engine.GET("/v1/call/handsome/:boy", func(c *gin.Context) {
		c.JSON(200, "call "+c.Param("boy"))
	})

	engine.Run(":29090")
}
