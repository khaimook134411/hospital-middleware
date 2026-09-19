package admin

import (
	"context"
	"errors"
	"log"

	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/util"
)

func (s *adminService) Login(ctx context.Context, username, password string) (string, *util.AppError) {
	staff, err := s.staffRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", util.ErrInvalidCredentials()
		}
		log.Printf("admin login: FindByUsername: %v", err)
		return "", util.ErrInternal()
	}
	if staff.Role != entity.RoleAdmin {
		return "", util.ErrInvalidCredentials()
	}
	if !util.CheckPassword(staff.PasswordHash, password) {
		return "", util.ErrInvalidCredentials()
	}
	token, err := util.GenerateToken(s.cfg, staff.ID, entity.RoleAdmin, nil, "")
	if err != nil {
		log.Printf("admin login: GenerateToken: %v", err)
		return "", util.ErrInternal()
	}
	return token, nil
}
