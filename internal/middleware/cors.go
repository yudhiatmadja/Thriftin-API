package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" { c.AbortWithStatus(204); return }
		c.Next()
	}
}
func Logger() gin.HandlerFunc { return gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
	return p.Method + " " + p.Path + " " + p.Latency.Truncate(time.Millisecond).String() + "\n"
})}
