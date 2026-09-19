package di

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/khaimook/hospital-middleware/client/his"
	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/di/database"
	"github.com/khaimook/hospital-middleware/di/server"
	"github.com/khaimook/hospital-middleware/handler"
	adminhandler "github.com/khaimook/hospital-middleware/handler/admin"
	patienthandler "github.com/khaimook/hospital-middleware/handler/patient"
	staffhandler "github.com/khaimook/hospital-middleware/handler/staff"
	hospitalrepo "github.com/khaimook/hospital-middleware/repository/hospital"
	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	staffrepo "github.com/khaimook/hospital-middleware/repository/staff"
	adminsvc "github.com/khaimook/hospital-middleware/service/admin"
	patientsvc "github.com/khaimook/hospital-middleware/service/patient"
	staffsvc "github.com/khaimook/hospital-middleware/service/staff"
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

	hospitalRepo := hospitalrepo.ProvideHospitalRepository(db)
	staffRepo := staffrepo.ProvideStaffRepository(db)
	patientRepo := patientrepo.ProvidePatientRepository(db)
	hisClient := his.ProvideHISClient(cfg)

	h := handler.Handlers{
		Admin:   adminhandler.NewHandler(adminsvc.ProvideAdminService(cfg, staffRepo), cfg),
		Staff:   staffhandler.NewHandler(staffsvc.ProvideStaffService(cfg, staffRepo, hospitalRepo), cfg),
		Patient: patienthandler.NewHandler(patientsvc.ProvidePatientService(patientRepo, hisClient)),
	}

	server.InitApiServer(cfg, db, h)
}
