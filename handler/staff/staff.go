package staff

import (
	"github.com/khaimook/hospital-middleware/di/config"
	svcstaff "github.com/khaimook/hospital-middleware/service/staff"
)

type Handler struct {
	svc svcstaff.StaffService
	cfg config.AppConfig
}

func NewHandler(svc svcstaff.StaffService, cfg config.AppConfig) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}
