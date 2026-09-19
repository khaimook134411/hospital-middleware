package staff_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
	mockshospital "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/repository/hospital"
	mocksstaff "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/repository/staff"
	"github.com/khaimook/hospital-middleware/util"
)

func staffEntity(password string) *entity.Staff {
	hash, _ := util.HashPassword(password)
	return &entity.Staff{ID: 5, Username: "nurse01", PasswordHash: hash, Role: entity.RoleStaff}
}

func TestLogin_Success(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	s := staffEntity("P@ssw0rd123")
	sr.EXPECT().FindByUsername(ctx, "nurse01").Return(s, nil)
	hr.EXPECT().FindByCode(ctx, "hospital-a").Return(hospitalA, nil)
	sr.EXPECT().HasHospital(ctx, uint(5), uint(1)).Return(true, nil)

	svc := newStaffService(sr, hr)
	token, appErr := svc.Login(ctx, "nurse01", "P@ssw0rd123", "hospital-a")

	require.Nil(t, appErr)
	assert.NotEmpty(t, token)

	claims, err := util.ParseToken(svcCfg, token)
	require.NoError(t, err)
	assert.Equal(t, entity.RoleStaff, claims.Role)
	require.NotNil(t, claims.HospitalID)
	assert.Equal(t, uint(1), *claims.HospitalID)
	assert.Equal(t, "hospital-a", claims.HospitalCode)
}

func TestLogin_TwoHospitals(t *testing.T) {
	for _, tc := range []struct {
		hospitalCode string
		hospital     *entity.Hospital
	}{
		{"hospital-a", hospitalA},
		{"hospital-b", hospitalB},
	} {
		t.Run("login at "+tc.hospitalCode, func(t *testing.T) {
			sr := mocksstaff.NewMockStaffRepository(t)
			hr := mockshospital.NewMockHospitalRepository(t)

			s := staffEntity("P@ssw0rd123")
			sr.EXPECT().FindByUsername(ctx, "nurse01").Return(s, nil)
			hr.EXPECT().FindByCode(ctx, tc.hospitalCode).Return(tc.hospital, nil)
			sr.EXPECT().HasHospital(ctx, uint(5), tc.hospital.ID).Return(true, nil)

			svc := newStaffService(sr, hr)
			token, appErr := svc.Login(ctx, "nurse01", "P@ssw0rd123", tc.hospitalCode)

			require.Nil(t, appErr)
			claims, err := util.ParseToken(svcCfg, token)
			require.NoError(t, err)
			assert.Equal(t, tc.hospital.ID, *claims.HospitalID)
			assert.Equal(t, tc.hospitalCode, claims.HospitalCode)
		})
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	s := staffEntity("correct")
	sr.EXPECT().FindByUsername(ctx, "nurse01").Return(s, nil)

	svc := newStaffService(sr, hr)
	_, appErr := svc.Login(ctx, "nurse01", "wrong", "hospital-a")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_UsernameNotFound(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	sr.EXPECT().FindByUsername(ctx, "nobody").Return(nil, gorm.ErrRecordNotFound)

	svc := newStaffService(sr, hr)
	_, appErr := svc.Login(ctx, "nobody", "P@ssw0rd123", "hospital-a")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_NoHospitalMapping(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	s := staffEntity("P@ssw0rd123")
	sr.EXPECT().FindByUsername(ctx, "nurse01").Return(s, nil)
	hr.EXPECT().FindByCode(ctx, "hospital-b").Return(hospitalB, nil)
	sr.EXPECT().HasHospital(ctx, uint(5), uint(2)).Return(false, nil)

	svc := newStaffService(sr, hr)
	_, appErr := svc.Login(ctx, "nurse01", "P@ssw0rd123", "hospital-b")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_AdminRoleRejected(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	hash, _ := util.HashPassword("adminpass")
	admin := &entity.Staff{ID: 1, Username: "admin", PasswordHash: hash, Role: entity.RoleAdmin}
	sr.EXPECT().FindByUsername(ctx, "admin").Return(admin, nil)

	svc := newStaffService(sr, hr)
	_, appErr := svc.Login(ctx, "admin", "adminpass", "hospital-a")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_HospitalNotFound(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	s := staffEntity("P@ssw0rd123")
	sr.EXPECT().FindByUsername(ctx, "nurse01").Return(s, nil)
	hr.EXPECT().FindByCode(ctx, "no-such").Return(nil, gorm.ErrRecordNotFound)

	svc := newStaffService(sr, hr)
	_, appErr := svc.Login(ctx, "nurse01", "P@ssw0rd123", "no-such")

	require.NotNil(t, appErr)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}

func TestLogin_DBError(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	sr.EXPECT().FindByUsername(ctx, "nurse01").Return(nil, errors.New("connection refused"))

	svc := newStaffService(sr, hr)
	_, appErr := svc.Login(ctx, "nurse01", "P@ssw0rd123", "hospital-a")

	require.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
}
