package patient

import (
	"context"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
)

type SearchParams struct {
	NationalID  string
	PassportID  string
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth string
	PhoneNumber string
	Email       string
	Limit       int
	Offset      int
}

type PatientRepository interface {
	Search(ctx context.Context, hospitalID uint, params SearchParams) ([]entity.Patient, error)
	Upsert(ctx context.Context, patient *entity.Patient) error
}

type patientRepository struct {
	db *gorm.DB
}

func ProvidePatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepository{db: db}
}
