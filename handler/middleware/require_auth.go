package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/util"
)

func RequireAuth(cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			util.WriteError(c, util.ErrUnauthorized())
			c.Abort()
			return
		}
		tokenStr := authHeader[len("Bearer "):]
		claims, err := util.ParseToken(cfg, tokenStr)
		if err != nil {
			util.WriteError(c, util.ErrUnauthorized())
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
