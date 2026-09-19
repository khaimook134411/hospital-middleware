package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/di"
	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/di/database"
	"github.com/khaimook/hospital-middleware/entity/migrator"
)

func main() {
	withMigrate := flag.Bool("with-migrate", false, "run database migration and exit")
	flag.Parse()

	if *withMigrate {
		cfg, err := config.GetConfig()
		if err != nil {
			log.Fatalf("config: %v", err)
		}
		gin.SetMode(cfg.GinMode)
		db, err := database.InitDatabase(cfg)
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		if err := migrator.Migrate(db, cfg); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		log.Println("migration completed")
		return
	}

	di.InitApplication()
}
