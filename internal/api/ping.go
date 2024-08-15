package api

import "github.com/gin-gonic/gin"

func Ping(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		c.JSON(200, Success("pong"))
	}

	router.GET("/ping", handler)
}
