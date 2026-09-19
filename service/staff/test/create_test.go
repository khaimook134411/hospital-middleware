package staff_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	mockshospital "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/repository/hospital"
	mocksstaff "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/repository/staff"
	svcstaff "github.com/khaimook/hospital-middleware/service/staff"
)

var svcCfg = config.AppConfig{
	JWTSecret: "test-secret-key-that-is-32-chars!!",
	JWTTTL:    time.Hour,
}

var ctx = context.Background()

func newStaffService(sr *mocksstaff.MockStaffRepository, hr *mockshospital.MockHospitalRepository) svcstaff.StaffService {
	return svcstaff.ProvideStaffService(svcCfg, sr, hr)
}

var hospitalA = &entity.Hospital{ID: 1, Code: "hospital-a", Name: "Hospital A"}
var hospitalB = &entity.Hospital{ID: 2, Code: "hospital-b", Name: "Hospital B"}

func TestCreate_Success(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	hr.EXPECT().FindByCode(ctx, "hospital-a").Return(hospitalA, nil)
	sr.EXPECT().CreateWithHospital(ctx, anyStaff(), uint(1)).Return(nil)

	svc := newStaffService(sr, hr)
	resp, appErr := svc.Create(ctx, "nurse01", "P@ssw0rd123", "hospital-a")

	require.Nil(t, appErr)
	require.NotNil(t, resp)
	assert.Equal(t, "nurse01", resp.Username)
	assert.Equal(t, "hospital-a", resp.Hospital)
}

func TestCreate_SuccessHospitalB(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	hr.EXPECT().FindByCode(ctx, "hospital-b").Return(hospitalB, nil)
	sr.EXPECT().CreateWithHospital(ctx, anyStaff(), uint(2)).Return(nil)

	svc := newStaffService(sr, hr)
	resp, appErr := svc.Create(ctx, "nurse02", "P@ssw0rd123", "hospital-b")

	require.Nil(t, appErr)
	assert.Equal(t, "hospital-b", resp.Hospital)
}

func TestCreate_HospitalNotFound(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	hr.EXPECT().FindByCode(ctx, "unknown").Return(nil, gorm.ErrRecordNotFound)

	svc := newStaffService(sr, hr)
	_, appErr := svc.Create(ctx, "nurse01", "P@ssw0rd123", "unknown")

	require.NotNil(t, appErr)
	assert.Equal(t, "HOSPITAL_NOT_FOUND", appErr.Code)
	assert.Equal(t, 404, appErr.Status)
}

func TestCreate_DuplicateUsername(t *testing.T) {
	for _, hospital := range []struct {
		code string
		id   uint
		h    *entity.Hospital
	}{
		{"hospital-a", 1, hospitalA},
		{"hospital-b", 2, hospitalB},
	} {
		t.Run("duplicate in "+hospital.code, func(t *testing.T) {
			sr := mocksstaff.NewMockStaffRepository(t)
			hr := mockshospital.NewMockHospitalRepository(t)

			hr.EXPECT().FindByCode(ctx, hospital.code).Return(hospital.h, nil)
			sr.EXPECT().CreateWithHospital(ctx, anyStaff(), hospital.id).
				Return(&pgconn.PgError{Code: "23505"})

			svc := newStaffService(sr, hr)
			_, appErr := svc.Create(ctx, "existing", "P@ssw0rd123", hospital.code)

			require.NotNil(t, appErr)
			assert.Equal(t, "USERNAME_TAKEN", appErr.Code)
			assert.Equal(t, 409, appErr.Status)
		})
	}
}

func TestCreate_DBErrorPropagated(t *testing.T) {
	sr := mocksstaff.NewMockStaffRepository(t)
	hr := mockshospital.NewMockHospitalRepository(t)

	hr.EXPECT().FindByCode(ctx, "hospital-a").Return(hospitalA, nil)
	sr.EXPECT().CreateWithHospital(ctx, anyStaff(), uint(1)).
		Return(errors.New("connection refused"))

	svc := newStaffService(sr, hr)
	_, appErr := svc.Create(ctx, "nurse01", "P@ssw0rd123", "hospital-a")

	require.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
}

func anyStaff() interface{} {
	return mock.MatchedBy(func(s *entity.Staff) bool { return s != nil })
}
