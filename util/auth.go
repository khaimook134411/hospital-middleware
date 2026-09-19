package util

import "github.com/gin-gonic/gin"

const claimsKey = "claims"

func GetClaims(c *gin.Context) *Claims {
	v, _ := c.Get(claimsKey)
	return v.(*Claims)
}
