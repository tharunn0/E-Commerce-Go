package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		reqId := uuid.New().String()
		c.Set("request_id", reqId)

		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		respSize := c.Writer.Size()

		var logFn func(msg string, fields ...zap.Field)
		switch {
		case status >= 500:
			logFn = log.Error
		case status >= 400:
			logFn = log.Warn
		default:
			logFn = log.Info
		}

		logFn("HTTP REQUEST",
			zap.String("request_id", reqId),
			zap.String("method", method),
			zap.String("service", "Requestlogger-Middlware"),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("response_size_bytes", respSize),
		)

	}
}
