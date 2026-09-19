package his

import (
	"context"
	"errors"

	"github.com/khaimook/hospital-middleware/di/config"
)

var (
	ErrPatientNotFound  = errors.New("his: patient not found")
	ErrHISNotConfigured = errors.New("his: hospital not configured")
)

type HISPatient struct {
	FirstNameTH  string `json:"first_name_th"`
	MiddleNameTH string `json:"middle_name_th"`
	LastNameTH   string `json:"last_name_th"`
	FirstNameEN  string `json:"first_name_en"`
	MiddleNameEN string `json:"middle_name_en"`
	LastNameEN   string `json:"last_name_en"`
	DateOfBirth  string `json:"date_of_birth"`
	PatientHN    string `json:"patient_hn"`
	NationalID   string `json:"national_id"`
	PassportID   string `json:"passport_id"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Gender       string `json:"gender"`
}

type HISClient interface {
	SearchPatient(ctx context.Context, hospitalCode, id string) (*HISPatient, error)
}

func ProvideHISClient(cfg config.AppConfig) HISClient {
	if cfg.HISMode == "mock" {
		return newHISMockClient()
	}
	return newHISHTTPClient(cfg)
}
