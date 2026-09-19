package staff

import (
	"context"

	"github.com/khaimook/hospital-middleware/entity"
)

func (r *staffRepository) HasHospital(ctx context.Context, staffID, hospitalID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.StaffHospital{}).
		Where("staff_id = ? AND hospital_id = ?", staffID, hospitalID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
