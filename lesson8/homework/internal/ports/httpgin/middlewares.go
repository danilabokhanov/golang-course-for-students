package httpgin

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func CustomLogger(c *gin.Context) {
	t := time.Now()
	path := c.Request.URL.Path

	c.Next()

	latency := time.Since(t)
	method := c.Request.Method
	status := c.Writer.Status()
	clientIP := c.ClientIP()
	bodySize := c.Writer.Size()

	log.Println("latency", latency, "method", method, "path", path, "status", status,
		"client_ip", clientIP, "body_size", bodySize)
}
