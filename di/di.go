package di

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/di/database"
	"github.com/khaimook/hospital-middleware/di/server"
)

func InitApplication() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	db, err := database.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	server.InitApiServer(cfg, db)
}
