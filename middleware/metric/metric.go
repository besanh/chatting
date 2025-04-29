package metric

import (
	"net/http"
	"time"

	"github.com/besanh/chatbot_gpt/config"
	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		statusCode := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		config.HttpRequestsTotal.WithLabelValues(c.Request.Method, path, http.StatusText(statusCode)).Inc()
		config.HttpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}
