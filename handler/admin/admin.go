package admin

import (
	"github.com/khaimook/hospital-middleware/di/config"
	svcadmin "github.com/khaimook/hospital-middleware/service/admin"
)

type Handler struct {
	svc svcadmin.AdminService
	cfg config.AppConfig
}

func NewHandler(svc svcadmin.AdminService, cfg config.AppConfig) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}
