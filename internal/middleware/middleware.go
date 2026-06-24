package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func LoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		fields := []any{
			"request_id", c.GetString(RequestIDKey),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}

		if c.Writer.Status() >= 500 {
			logger.Error("request completed", fields...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}

func CustomRecovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if c.IsAborted() {
					return
				}
				logger.Error("panic recovered",
					"request_id", c.GetString(RequestIDKey),
					"error", err,
				)
				c.AbortWithStatusJSON(500, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
