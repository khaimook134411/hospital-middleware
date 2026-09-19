package patient

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/khaimook/hospital-middleware/client/his"
	"github.com/khaimook/hospital-middleware/entity"
	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	"github.com/khaimook/hospital-middleware/util"
)

func (s *patientService) Search(ctx context.Context, hospitalID uint, hospitalCode string, params patientrepo.SearchParams) ([]entity.PatientResponse, *util.AppError) {
	patients, err := s.patientRepo.Search(ctx, hospitalID, params)
	if err != nil {
		log.Printf("patient search: %v", err)
		return nil, util.ErrInternal()
	}

	if len(patients) == 0 && (params.NationalID != "" || params.PassportID != "") {
		id := params.NationalID
		if id == "" {
			id = params.PassportID
		}

		hisPatient, hisErr := s.hisClient.SearchPatient(ctx, hospitalCode, id)
		if hisErr != nil {
			if errors.Is(hisErr, his.ErrPatientNotFound) {
				return []entity.PatientResponse{}, nil
			}
			if errors.Is(hisErr, his.ErrHISNotConfigured) {
				log.Printf("HIS not configured for hospital %q: %v", hospitalCode, hisErr)
				return []entity.PatientResponse{}, nil
			}
			log.Printf("HIS error for hospital %q: %v", hospitalCode, hisErr)
			return nil, util.ErrHISUnavailable()
		}

		patient := hisPatientToEntity(hisPatient, hospitalID)
		if upsertErr := s.patientRepo.Upsert(ctx, patient); upsertErr != nil {
			log.Printf("patient upsert: %v", upsertErr)
			return nil, util.ErrInternal()
		}

		patients, err = s.patientRepo.Search(ctx, hospitalID, params)
		if err != nil {
			log.Printf("patient re-search: %v", err)
			return nil, util.ErrInternal()
		}
	}

	responses := make([]entity.PatientResponse, 0, len(patients))
	for _, p := range patients {
		responses = append(responses, p.ToResponse())
	}
	return responses, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func hisPatientToEntity(h *his.HISPatient, hospitalID uint) *entity.Patient {
	var dob *entity.Date
	if h.DateOfBirth != "" {
		t, _ := time.Parse("2006-01-02", h.DateOfBirth)
		d := entity.Date{Time: t}
		dob = &d
	}

	return &entity.Patient{
		HospitalID:   hospitalID,
		PatientHN:    h.PatientHN,
		NationalID:   strPtr(h.NationalID),
		PassportID:   strPtr(h.PassportID),
		FirstNameTH:  strPtr(h.FirstNameTH),
		MiddleNameTH: strPtr(h.MiddleNameTH),
		LastNameTH:   strPtr(h.LastNameTH),
		FirstNameEN:  strPtr(h.FirstNameEN),
		MiddleNameEN: strPtr(h.MiddleNameEN),
		LastNameEN:   strPtr(h.LastNameEN),
		DateOfBirth:  dob,
		PhoneNumber:  strPtr(h.PhoneNumber),
		Email:        strPtr(h.Email),
		Gender:       strPtr(h.Gender),
	}
}
