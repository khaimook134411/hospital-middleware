package staff

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/khaimook/hospital-middleware/di/config"
	hospitalrepo "github.com/khaimook/hospital-middleware/repository/hospital"
	staffrepo "github.com/khaimook/hospital-middleware/repository/staff"
	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/util"
)

type StaffService interface {
	Create(ctx context.Context, username, password, hospitalCode string) (*entity.StaffResponse, *util.AppError)
	Login(ctx context.Context, username, password, hospitalCode string) (string, *util.AppError)
}

type staffService struct {
	cfg          config.AppConfig
	staffRepo    staffrepo.StaffRepository
	hospitalRepo hospitalrepo.HospitalRepository
}

func ProvideStaffService(
	cfg config.AppConfig,
	staffRepo staffrepo.StaffRepository,
	hospitalRepo hospitalrepo.HospitalRepository,
) StaffService {
	return &staffService{cfg: cfg, staffRepo: staffRepo, hospitalRepo: hospitalRepo}
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
