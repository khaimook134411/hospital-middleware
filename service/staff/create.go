package staff

import (
	"context"
	"errors"
	"log"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/util"
)

func (s *staffService) Create(ctx context.Context, username, password, hospitalCode string) (*entity.StaffResponse, *util.AppError) {
	hospital, err := s.hospitalRepo.FindByCode(ctx, hospitalCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrHospitalNotFound()
		}
		log.Printf("staff create: FindByCode: %v", err)
		return nil, util.ErrInternal()
	}

	hash, err := util.HashPassword(password)
	if err != nil {
		log.Printf("staff create: HashPassword: %v", err)
		return nil, util.ErrInternal()
	}

	staff := &entity.Staff{
		Username:     username,
		PasswordHash: hash,
		Role:         entity.RoleStaff,
	}
	if err := s.staffRepo.CreateWithHospital(ctx, staff, hospital.ID); err != nil {
		if isDuplicateKey(err) {
			return nil, util.ErrUsernameTaken()
		}
		log.Printf("staff create: CreateWithHospital: %v", err)
		return nil, util.ErrInternal()
	}

	resp := staff.ToResponse(hospitalCode)
	return &resp, nil
}
