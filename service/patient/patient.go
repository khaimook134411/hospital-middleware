package patient

import (
	"context"

	"github.com/khaimook/hospital-middleware/client/his"
	"github.com/khaimook/hospital-middleware/entity"
	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	"github.com/khaimook/hospital-middleware/util"
)

type PatientService interface {
	Search(ctx context.Context, hospitalID uint, hospitalCode string, params patientrepo.SearchParams) ([]entity.PatientResponse, *util.AppError)
}

type patientService struct {
	patientRepo patientrepo.PatientRepository
	hisClient   his.HISClient
}

func ProvidePatientService(patientRepo patientrepo.PatientRepository, hisClient his.HISClient) PatientService {
	return &patientService{patientRepo: patientRepo, hisClient: hisClient}
}
