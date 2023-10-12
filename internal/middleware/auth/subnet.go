package auth

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewSubnetChecker(trustedSubnet string, logger *zap.SugaredLogger) gin.HandlerFunc {
	_, netMask, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logger.Warnf("cannot parse trusted subnet, internal service unavaliable: %w", err)
	}

	return func(c *gin.Context) {
		if netMask == nil {
			logger.Error("trusted subnet is not defined")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		realIP := c.GetHeader("X-Real-IP")
		if realIP == "" {
			logger.Error("internal request: empty X-Real-IP")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		ipAddr := net.ParseIP(realIP)
		if ipAddr == nil {
			logger.Error("internal request: error parsing X-Real-IP")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !netMask.Contains(ipAddr) {
			logger.Errorf("internal request: unauthorized request denied: ip %s", ipAddr)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
