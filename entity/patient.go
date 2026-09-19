package entity

import "time"

type Patient struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	HospitalID   uint    `gorm:"not null"`
	PatientHN    string  `gorm:"not null"`
	NationalID   *string
	PassportID   *string
	FirstNameTH  *string
	MiddleNameTH *string
	LastNameTH   *string
	FirstNameEN  *string
	MiddleNameEN *string
	LastNameEN   *string
	DateOfBirth  *Date   `gorm:"type:date"`
	PhoneNumber  *string
	Email        *string
	Gender       *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PatientResponse struct {
	PatientHN    string  `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	FirstNameTH  *string `json:"first_name_th"`
	MiddleNameTH *string `json:"middle_name_th"`
	LastNameTH   *string `json:"last_name_th"`
	FirstNameEN  *string `json:"first_name_en"`
	MiddleNameEN *string `json:"middle_name_en"`
	LastNameEN   *string `json:"last_name_en"`
	DateOfBirth  *Date   `json:"date_of_birth"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
	Gender       *string `json:"gender"`
}

func (p *Patient) ToResponse() PatientResponse {
	return PatientResponse{
		PatientHN:    p.PatientHN,
		NationalID:   p.NationalID,
		PassportID:   p.PassportID,
		FirstNameTH:  p.FirstNameTH,
		MiddleNameTH: p.MiddleNameTH,
		LastNameTH:   p.LastNameTH,
		FirstNameEN:  p.FirstNameEN,
		MiddleNameEN: p.MiddleNameEN,
		LastNameEN:   p.LastNameEN,
		DateOfBirth:  p.DateOfBirth,
		PhoneNumber:  p.PhoneNumber,
		Email:        p.Email,
		Gender:       p.Gender,
	}
}
