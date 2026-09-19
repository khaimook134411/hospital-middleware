package patient

import (
	"context"
	"strings"

	"github.com/khaimook/hospital-middleware/entity"
)

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *patientRepository) Search(ctx context.Context, hospitalID uint, params SearchParams) ([]entity.Patient, error) {
	q := r.db.WithContext(ctx).Where("hospital_id = ?", hospitalID)

	if params.NationalID != "" {
		q = q.Where("national_id = ?", params.NationalID)
	}
	if params.PassportID != "" {
		q = q.Where("passport_id = ?", params.PassportID)
	}
	if params.PhoneNumber != "" {
		q = q.Where("phone_number = ?", params.PhoneNumber)
	}
	if params.DateOfBirth != "" {
		q = q.Where("date_of_birth = ?", params.DateOfBirth)
	}
	if params.Email != "" {
		q = q.Where("LOWER(email) = LOWER(?)", params.Email)
	}
	if params.FirstName != "" {
		pattern := "%" + escapeLike(params.FirstName) + "%"
		q = q.Where("(first_name_th ILIKE ? OR first_name_en ILIKE ?)", pattern, pattern)
	}
	if params.MiddleName != "" {
		pattern := "%" + escapeLike(params.MiddleName) + "%"
		q = q.Where("(middle_name_th ILIKE ? OR middle_name_en ILIKE ?)", pattern, pattern)
	}
	if params.LastName != "" {
		pattern := "%" + escapeLike(params.LastName) + "%"
		q = q.Where("(last_name_th ILIKE ? OR last_name_en ILIKE ?)", pattern, pattern)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	var patients []entity.Patient
	err := q.Order("id ASC").Limit(limit).Offset(params.Offset).Find(&patients).Error
	return patients, err
}
