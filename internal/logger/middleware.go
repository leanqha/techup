package logger

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GinLoggerMiddleware logs each HTTP request using zerolog
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // обработать запрос

		latency := time.Since(start)
		status := c.Writer.Status()

		requestID, _ := c.Get("request_id")
		errMsg := ""
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		Log.Info().
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Str("client_ip", c.ClientIP()).
			Int("status", status).
			Dur("latency", latency).
			Str("request_id", requestID.(string)).
			Str("error", errMsg).
			Msg("request completed")
	}
}

func LogSQLError(err error, query string, args ...interface{}) {
	if err != nil {
		Log.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("database query failed")
	}
}

func RecoveryLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				Log.Error().
					Interface("panic", r).
					Str("path", c.Request.URL.Path).
					Msg("panic recovered")

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}
