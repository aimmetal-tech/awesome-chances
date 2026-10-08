package router

import (
	"log/slog"
	"time"

	"awesome-chances/backend/internal/handler"
	"awesome-chances/backend/internal/security"
	"github.com/gin-gonic/gin"
)

// requestLog 只记录方法/路径/状态和耗时，不记录请求体、密码、Cookie 或查询参数。
func requestLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := security.RandomID()
		c.Header("X-Request-ID", id)
		c.Header("Cache-Control", "no-store")
		c.Header("X-Content-Type-Options", "nosniff")
		start := time.Now()
		defer func() {
			if recover() != nil {
				logger.Error("request panic recovered", "requestId", id)
				if !c.Writer.Written() {
					c.JSON(500, handler.RouteError("internal_error", "服务暂时无法处理请求"))
				}
				c.Abort()
			}
			logger.Info("http request", "requestId", id, "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "durationMs", time.Since(start).Milliseconds())
		}()
		c.Next()
	}
}
