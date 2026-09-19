package staff

import (
	"context"

	"github.com/khaimook/hospital-middleware/entity"
)

func (r *staffRepository) FindByUsername(ctx context.Context, username string) (*entity.Staff, error) {
	var s entity.Staff
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
