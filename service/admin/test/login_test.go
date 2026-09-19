package admin_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	mocksstaff "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/repository/staff"
	svcadmin "github.com/khaimook/hospital-middleware/service/admin"
	"github.com/khaimook/hospital-middleware/util"
)

var svcCfg = config.AppConfig{
	JWTSecret: "test-secret-key-that-is-32-chars!!",
	JWTTTL:    time.Hour,
}

func newService(repo *mocksstaff.MockStaffRepository) svcadmin.AdminService {
	return svcadmin.ProvideAdminService(svcCfg, repo)
}

func adminStaff(password string) *entity.Staff {
	hash, _ := util.HashPassword(password)
	return &entity.Staff{ID: 1, Username: "admin", PasswordHash: hash, Role: entity.RoleAdmin}
}

func TestLogin_Success(t *testing.T) {
	repo := mocksstaff.NewMockStaffRepository(t)
	s := adminStaff("secret123")
	repo.EXPECT().FindByUsername(context.Background(), "admin").Return(s, nil)

	svc := newService(repo)
	token, appErr := svc.Login(context.Background(), "admin", "secret123")

	require.Nil(t, appErr)
	assert.NotEmpty(t, token)

	claims, err := util.ParseToken(svcCfg, token)
	require.NoError(t, err)
	assert.Equal(t, "1", claims.Subject)
	assert.Equal(t, entity.RoleAdmin, claims.Role)
}

func TestLogin_UnknownUsername(t *testing.T) {
	repo := mocksstaff.NewMockStaffRepository(t)
	repo.EXPECT().FindByUsername(context.Background(), "nobody").Return(nil, gorm.ErrRecordNotFound)

	svc := newService(repo)
	_, appErr := svc.Login(context.Background(), "nobody", "pass")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := mocksstaff.NewMockStaffRepository(t)
	s := adminStaff("correct")
	repo.EXPECT().FindByUsername(context.Background(), "admin").Return(s, nil)

	svc := newService(repo)
	_, appErr := svc.Login(context.Background(), "admin", "wrong")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_StaffRoleReturnsInvalidCredentials(t *testing.T) {
	repo := mocksstaff.NewMockStaffRepository(t)
	hash, _ := util.HashPassword("pass")
	s := &entity.Staff{ID: 2, Username: "staff1", PasswordHash: hash, Role: entity.RoleStaff}
	repo.EXPECT().FindByUsername(context.Background(), "staff1").Return(s, nil)

	svc := newService(repo)
	_, appErr := svc.Login(context.Background(), "staff1", "pass")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_DBError(t *testing.T) {
	repo := mocksstaff.NewMockStaffRepository(t)
	repo.EXPECT().FindByUsername(context.Background(), "admin").Return(nil, errors.New("connection refused"))

	svc := newService(repo)
	_, appErr := svc.Login(context.Background(), "admin", "pass")

	require.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
}
