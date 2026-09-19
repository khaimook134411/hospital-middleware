package patient

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/khaimook/hospital-middleware/entity"
)

func (r *patientRepository) Upsert(ctx context.Context, patient *entity.Patient) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hospital_id"}, {Name: "patient_hn"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"national_id", "passport_id",
				"first_name_th", "middle_name_th", "last_name_th",
				"first_name_en", "middle_name_en", "last_name_en",
				"date_of_birth", "phone_number", "email", "gender",
				"updated_at",
			}),
		}).
		Create(patient).Error
}
