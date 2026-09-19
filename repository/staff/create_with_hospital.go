package staff

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
)

func (r *staffRepository) CreateWithHospital(ctx context.Context, staff *entity.Staff, hospitalID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(staff).Error; err != nil {
			return fmt.Errorf("create staff: %w", err)
		}
		mapping := entity.StaffHospital{
			StaffID:    staff.ID,
			HospitalID: hospitalID,
		}
		if err := tx.Create(&mapping).Error; err != nil {
			return fmt.Errorf("create staff_hospital: %w", err)
		}
		return nil
	})
}
