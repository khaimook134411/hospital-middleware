package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func WriteSuccessWithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}

func WriteError(c *gin.Context, appErr *AppError) {
	c.JSON(appErr.Status, gin.H{"error": gin.H{
		"code":    appErr.Code,
		"message": appErr.Message,
	}})
}
