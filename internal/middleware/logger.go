package middleware

import (
	"time"
	"upcycle-hub/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start)
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("cost", cost),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
			logger.Warn("http request", fields...)
		} else {
			logger.Info("http request", fields...)
		}
	}
}
