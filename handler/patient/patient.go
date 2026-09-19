package patient

import (
	patientsvc "github.com/khaimook/hospital-middleware/service/patient"
)

type Handler struct {
	svc patientsvc.PatientService
}

func NewHandler(svc patientsvc.PatientService) *Handler {
	return &Handler{svc: svc}
}
