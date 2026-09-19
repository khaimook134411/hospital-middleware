package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	adminhandler "github.com/khaimook/hospital-middleware/handler/admin"
	"github.com/khaimook/hospital-middleware/handler/middleware"
	patienthandler "github.com/khaimook/hospital-middleware/handler/patient"
	staffhandler "github.com/khaimook/hospital-middleware/handler/staff"
)

type Handlers struct {
	Admin   *adminhandler.Handler
	Staff   *staffhandler.Handler
	Patient *patienthandler.Handler
}

func InitRouter(r *gin.Engine, db *gorm.DB, cfg config.AppConfig, h Handlers) {
	r.GET("/health", healthHandler(db))

	r.POST("/admin/login", h.Admin.Login)

	staffGroup := r.Group("/staff")
	{
		staffGroup.POST("/create",
			middleware.RequireAuth(cfg),
			middleware.RequireRole(entity.RoleAdmin),
			h.Staff.Create,
		)
		staffGroup.POST("/login", h.Staff.Login)
	}

	patientGroup := r.Group("/patient")
	{
		patientGroup.GET("/search",
			middleware.RequireAuth(cfg),
			middleware.RequireRole(entity.RoleStaff),
			h.Patient.Search,
		)
	}
}

func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database unavailable"})
			return
		}
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
