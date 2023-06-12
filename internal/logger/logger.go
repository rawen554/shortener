package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() (gin.HandlerFunc, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		method := c.Request.Method

		t := time.Now()
		c.Next()
		duration := time.Since(t)

		sugar.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"Status", c.Writer.Status(),
			"Size", c.Writer.Size(),
		)
	}, nil
}
