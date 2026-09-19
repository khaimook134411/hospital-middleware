package hospital

import (
	"context"

	"github.com/khaimook/hospital-middleware/entity"
)

func (r *hospitalRepository) FindByCode(ctx context.Context, code string) (*entity.Hospital, error) {
	var h entity.Hospital
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}
