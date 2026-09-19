package staff

import (
	"context"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
)

type StaffRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.Staff, error)
	CreateWithHospital(ctx context.Context, staff *entity.Staff, hospitalID uint) error
	HasHospital(ctx context.Context, staffID, hospitalID uint) (bool, error)
}

type staffRepository struct {
	db *gorm.DB
}

func ProvideStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}
