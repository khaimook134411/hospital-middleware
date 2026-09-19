package hospital

import (
	"context"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
)

type HospitalRepository interface {
	FindByCode(ctx context.Context, code string) (*entity.Hospital, error)
}

type hospitalRepository struct {
	db *gorm.DB
}

func ProvideHospitalRepository(db *gorm.DB) HospitalRepository {
	return &hospitalRepository{db: db}
}
