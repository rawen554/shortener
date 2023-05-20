package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	return func(c *gin.Context) {
		t := time.Now()
		uri := c.Request.RequestURI
		method := c.Request.Method

		c.Next()

		duration := time.Since(t)
		sugar.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"Status", c.Writer.Status(),
			"Size", c.Writer.Size(),
		)
	}
}
