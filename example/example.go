package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/huzhongqing/ginprom"
	"github.com/prometheus/client_golang/prometheus"
	"os"
)

func main() {
	engine := gin.New()

	cfg := ginprom.Config{
		FixedPath: []string{"/v1/hello"},
		ParamsPath: map[string]int{
			"/v1/hello/:name":  3,
			"/hello/:sex/call": 2,
		},
	}

	gp := ginprom.New(engine, &cfg)

	gp.DefaultRegister("test")

	gp.Use(engine)

	// 执行默认的
	engine.Use(gp.DefaultHandlerFunc())

	// 自定义自己想要的结构
	iMetric := ginprom.Metric{
		ID:          "PV",
		Name:        "requests_pv",
		Description: "requests pv",
		Type:        "counter",
	}

	if err := gp.AddMetrics(&iMetric, "test"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	handlerF := func(v *ginprom.GinPrometheus) gin.HandlerFunc {
		return func(c *gin.Context) {
			pv, ok := v.MetricsMap.Load("PV")
			if ok {
				pv.(ginprom.Metric).Collector.(prometheus.Counter).Inc()
			}
			c.Next()
		}
	}

	engine.Use(handlerF(gp))

	engine.GET("/v1/hello", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	engine.GET("/v1/hello/:name", func(c *gin.Context) {
		c.JSON(200, "hello"+c.Param("name"))
	})

	engine.GET("/hello/:sex/call", func(c *gin.Context) {
		c.JSON(200, "I am"+c.Param("sex"))
	})

	engine.Run(":29090")
}
