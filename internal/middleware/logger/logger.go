// Модуль логирования запросов.
package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger Получение middleware функции, которая будет логгировать входящие запросы.
func Logger(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		method := c.Request.Method

		t := time.Now()
		c.Next()
		duration := time.Since(t)

		logger.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"Status", c.Writer.Status(),
			"Size", c.Writer.Size(),
		)
	}
}
