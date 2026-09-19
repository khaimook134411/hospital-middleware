package admin

import (
	"context"

	"github.com/khaimook/hospital-middleware/di/config"
	staffrepo "github.com/khaimook/hospital-middleware/repository/staff"
	"github.com/khaimook/hospital-middleware/util"
)

type AdminService interface {
	Login(ctx context.Context, username, password string) (string, *util.AppError)
}

type adminService struct {
	cfg       config.AppConfig
	staffRepo staffrepo.StaffRepository
}

func ProvideAdminService(cfg config.AppConfig, staffRepo staffrepo.StaffRepository) AdminService {
	return &adminService{cfg: cfg, staffRepo: staffRepo}
}
