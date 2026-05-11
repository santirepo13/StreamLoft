package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"streamloft-api/internal/logger"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log := logger.Get()
		if status >= 500 {
			log.Error().
				Int("status", status).
				Str("method", method).
				Str("path", path).
				Dur("latency", latency).
				Msg("request completed with error")
		} else if status >= 400 {
			log.Warn().
				Int("status", status).
				Str("method", method).
				Str("path", path).
				Dur("latency", latency).
				Msg("request completed with client error")
		} else {
			log.Info().
				Int("status", status).
				Str("method", method).
				Str("path", path).
				Dur("latency", latency).
				Msg("request completed")
		}
	}
}