package staff

import (
	"context"
	"errors"
	"log"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/util"
)

func (s *staffService) Login(ctx context.Context, username, password, hospitalCode string) (string, *util.AppError) {
	staff, err := s.staffRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", util.ErrInvalidCredentials()
		}
		log.Printf("staff login: FindByUsername: %v", err)
		return "", util.ErrInternal()
	}

	// admin ต้องใช้ /admin/login แทน — ตอบ INVALID_CREDENTIALS เพื่อไม่ให้เดาได้
	if staff.Role != entity.RoleStaff {
		return "", util.ErrInvalidCredentials()
	}

	if !util.CheckPassword(staff.PasswordHash, password) {
		return "", util.ErrInvalidCredentials()
	}

	hospital, err := s.hospitalRepo.FindByCode(ctx, hospitalCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", util.ErrInvalidCredentials()
		}
		log.Printf("staff login: FindByCode: %v", err)
		return "", util.ErrInternal()
	}

	hasMapping, err := s.staffRepo.HasHospital(ctx, staff.ID, hospital.ID)
	if err != nil {
		log.Printf("staff login: HasHospital: %v", err)
		return "", util.ErrInternal()
	}
	if !hasMapping {
		return "", util.ErrInvalidCredentials()
	}

	token, err := util.GenerateToken(s.cfg, staff.ID, entity.RoleStaff, &hospital.ID, hospital.Code)
	if err != nil {
		log.Printf("staff login: GenerateToken: %v", err)
		return "", util.ErrInternal()
	}
	return token, nil
}
