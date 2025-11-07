package logger

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinLoggerMiddleware logs each HTTP request using zerolog
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// After request
		latency := time.Since(start)
		status := c.Writer.Status()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		Log.Info().
			Str("method", method).
			Str("path", path).
			Str("client_ip", clientIP).
			Int("status", status).
			Dur("latency", latency).
			Msg("HTTP request")
	}
}
