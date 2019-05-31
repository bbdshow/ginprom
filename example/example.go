package main

import (
	"github.com/gin-gonic/gin"
	"github.com/huzhongqing/ginprom"
)

func main() {
	engine := gin.New()

	cfg := ginprom.Config{
		FixedPath: []string{"/v1/hello"},
		ParamsPath: map[string]int{
			"/v1/hello/:name": 3,
			"/hello/:sex/call": 2,
		},
	}

	p := ginprom.New(engine, &cfg)

	p.Use(engine)

	engine.Use(p.HandlerFunc())

	engine.GET("/v1/hello", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	engine.GET("/v1/hello/:name", func(c *gin.Context){
		c.JSON(200, "hello" + c.Param("name"))
	})

	engine.GET("/hello/:sex/call", func(c *gin.Context) {
		c.JSON(200, "I am" + c.Param("sex"))
	})

	engine.Run(":29090")
}
