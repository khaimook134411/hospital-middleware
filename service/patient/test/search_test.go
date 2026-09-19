package patient_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/khaimook/hospital-middleware/client/his"
	"github.com/khaimook/hospital-middleware/entity"
	mockshis "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/client/his"
	mockspatient "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/repository/patient"
	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	patientsvc "github.com/khaimook/hospital-middleware/service/patient"
)

var ctx = context.Background()

const hospitalID = uint(1)
const hospitalCode = "hospital-a"

func newSvc(repo *mockspatient.MockPatientRepository, hisClient *mockshis.MockHISClient) patientsvc.PatientService {
	return patientsvc.ProvidePatientService(repo, hisClient)
}

func strPtr(s string) *string { return &s }

func makePatient(hn, nationalID string) entity.Patient {
	dob := entity.Date{Time: time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC)}
	gender := "M"
	return entity.Patient{
		ID:          1,
		HospitalID:  hospitalID,
		PatientHN:   hn,
		NationalID:  strPtr(nationalID),
		FirstNameEN: strPtr("Somchai"),
		LastNameEN:  strPtr("Jaidee"),
		DateOfBirth: &dob,
		Gender:      &gender,
	}
}

func TestSearch_DBHasResults(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}
	patient := makePatient("HNA001", "1234567890123")
	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{patient}, nil)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	require.Len(t, results, 1)
	assert.Equal(t, "HNA001", results[0].PatientHN)
	assert.Equal(t, strPtr("1234567890123"), results[0].NationalID)
}

func TestSearch_MultipleParams(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{FirstName: "Somchai", LastName: "Jaidee", Limit: 20}
	patient := makePatient("HNA001", "1234567890123")
	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{patient}, nil)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	require.Len(t, results, 1)
}

func TestSearch_NoParams(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{Limit: 20}
	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	assert.Empty(t, results)
}

func TestSearch_HISFallback_NationalID(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}
	patient := makePatient("HNA001", "1234567890123")

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil).Once()
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "1234567890123").Return(&his.HISPatient{
		PatientHN:   "HNA001",
		NationalID:  "1234567890123",
		FirstNameEN: "Somchai",
		LastNameEN:  "Jaidee",
		DateOfBirth: "1990-05-20",
		Gender:      "M",
	}, nil)
	repo.EXPECT().Upsert(ctx, mock.MatchedBy(func(p *entity.Patient) bool {
		return p != nil && p.PatientHN == "HNA001" && p.HospitalID == hospitalID
	})).Return(nil)
	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{patient}, nil).Once()

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	require.Len(t, results, 1)
	assert.Equal(t, "HNA001", results[0].PatientHN)
}

func TestSearch_HISFallback_PassportID(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{PassportID: "AA1234567", Limit: 20}
	patient := makePatient("HNA003", "")

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil).Once()
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "AA1234567").Return(&his.HISPatient{
		PatientHN:  "HNA003",
		PassportID: "AA1234567",
		Gender:     "M",
	}, nil)
	repo.EXPECT().Upsert(ctx, mock.MatchedBy(func(p *entity.Patient) bool {
		return p != nil && p.PatientHN == "HNA003"
	})).Return(nil)
	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{patient}, nil).Once()

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	require.Len(t, results, 1)
}

func TestSearch_HISFallback_PrefersNationalID(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", PassportID: "AA1234567", Limit: 20}

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil).Once()
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "1234567890123").Return(nil, his.ErrPatientNotFound)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	assert.Empty(t, results)
}

func TestSearch_HISFallback_404(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "0000000000000", Limit: 20}

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil)
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "0000000000000").Return(nil, his.ErrPatientNotFound)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	assert.Empty(t, results)
}

func TestSearch_HISFallback_NotConfigured(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil)
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "1234567890123").Return(nil, his.ErrHISNotConfigured)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.Nil(t, appErr)
	assert.Empty(t, results)
}

func TestSearch_HISFallback_Unavailable(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil)
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "1234567890123").Return(nil, errors.New("connection timeout"))

	svc := newSvc(repo, hisClient)
	_, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.NotNil(t, appErr)
	assert.Equal(t, "HIS_UNAVAILABLE", appErr.Code)
	assert.Equal(t, 502, appErr.Status)
}

func TestSearch_DBError(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}
	repo.EXPECT().Search(ctx, hospitalID, params).Return(nil, errors.New("connection refused"))

	svc := newSvc(repo, hisClient)
	_, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
}

func TestSearch_UpsertError(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil)
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "1234567890123").Return(&his.HISPatient{
		PatientHN: "HNA001", Gender: "M",
	}, nil)
	repo.EXPECT().Upsert(ctx, mock.Anything).Return(errors.New("db write error"))

	svc := newSvc(repo, hisClient)
	_, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)

	require.NotNil(t, appErr)
	assert.Equal(t, "INTERNAL_ERROR", appErr.Code)
}

func TestSearch_HISEmptyStringToNil(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	params := patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20}

	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil).Once()
	hisClient.EXPECT().SearchPatient(ctx, hospitalCode, "1234567890123").Return(&his.HISPatient{
		PatientHN:  "HNA001",
		NationalID: "1234567890123",
	}, nil)
	repo.EXPECT().Upsert(ctx, mock.MatchedBy(func(p *entity.Patient) bool {
		return p.FirstNameTH == nil && p.LastNameEN == nil && p.Email == nil
	})).Return(nil)
	repo.EXPECT().Search(ctx, hospitalID, params).Return([]entity.Patient{}, nil).Once()

	svc := newSvc(repo, hisClient)
	_, appErr := svc.Search(ctx, hospitalID, hospitalCode, params)
	require.Nil(t, appErr)
}

func TestSearch_HospitalIsolation(t *testing.T) {
	repo := mockspatient.NewMockPatientRepository(t)
	hisClient := mockshis.NewMockHISClient(t)

	const hospitalBID = uint(2)
	params := patientrepo.SearchParams{LastName: "Jaidee", Limit: 20}

	// hospital B's repo call returns nothing for hospital B's isolation
	repo.EXPECT().Search(ctx, hospitalBID, params).Return([]entity.Patient{}, nil)

	svc := newSvc(repo, hisClient)
	results, appErr := svc.Search(ctx, hospitalBID, "hospital-b", params)

	require.Nil(t, appErr)
	assert.Empty(t, results)
}
