package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/handler"
	"github.com/khaimook/hospital-middleware/handler/middleware"
)

func BuildApp(db *gorm.DB, cfg config.AppConfig, h handler.Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	handler.InitRouter(r, db, cfg, h)
	return r
}

func InitApiServer(cfg config.AppConfig, db *gorm.DB, h handler.Handlers) {
	r := BuildApp(db, cfg, h)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.AppPort),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	log.Printf("server listening on :%s", cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
