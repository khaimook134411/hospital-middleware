package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/util"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := util.GetClaims(c)
		if claims.Role != role {
			util.WriteError(c, util.ErrForbidden())
			c.Abort()
			return
		}
		c.Next()
	}
}
