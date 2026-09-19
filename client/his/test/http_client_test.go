package his_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khaimook/hospital-middleware/client/his"
	"github.com/khaimook/hospital-middleware/di/config"
)

func newHTTPCfg(baseURL string) config.AppConfig {
	return config.AppConfig{
		HISMode:     "http",
		HISTimeout:  5 * time.Second,
		HISBaseURLs: config.HISBaseURLMap{"hospital-a": baseURL},
	}
}

// ✅ 200 OK → returns patient, ตรวจ method และ path
func TestSearchPatient_200(t *testing.T) {
	want := his.HISPatient{
		PatientHN:   "HNA001",
		NationalID:  "1234567890123",
		FirstNameTH: "สมชาย",
		LastNameTH:  "ใจดี",
		DateOfBirth: "1990-05-20",
		Gender:      "M",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/patient/search/1234567890123", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	got, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want.PatientHN, got.PatientHN)
	assert.Equal(t, want.NationalID, got.NationalID)
	assert.Equal(t, want.FirstNameTH, got.FirstNameTH)
	assert.Equal(t, "M", got.Gender)
}

// ✅ gender หรือ DOB เป็น empty string → ผ่านการ validate (ไม่ error)
func TestSearchPatient_EmptyOptionalFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(his.HISPatient{PatientHN: "HNA999"})
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	got, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.NoError(t, err)
	assert.Equal(t, "HNA999", got.PatientHN)
}

// ❌ 404 → ErrPatientNotFound
func TestSearchPatient_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	_, err := client.SearchPatient(context.Background(), "hospital-a", "0000000000000")

	require.ErrorIs(t, err, his.ErrPatientNotFound)
}

// ❌ 500 → error (non-ErrPatientNotFound, non-ErrHISNotConfigured)
func TestSearchPatient_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	_, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.Error(t, err)
	assert.NotErrorIs(t, err, his.ErrPatientNotFound)
	assert.NotErrorIs(t, err, his.ErrHISNotConfigured)
}

// ❌ timeout → error
func TestSearchPatient_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.AppConfig{
		HISMode:     "http",
		HISTimeout:  10 * time.Millisecond,
		HISBaseURLs: config.HISBaseURLMap{"hospital-a": srv.URL},
	}
	client := his.ProvideHISClient(cfg)
	_, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.Error(t, err)
}

// ❌ JSON ผิดรูปแบบ → error
func TestSearchPatient_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{not-valid-json"))
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	_, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.Error(t, err)
}

// ❌ hospital code ไม่อยู่ใน config → ErrHISNotConfigured
func TestSearchPatient_HospitalNotConfigured(t *testing.T) {
	cfg := config.AppConfig{
		HISMode:     "http",
		HISTimeout:  5 * time.Second,
		HISBaseURLs: config.HISBaseURLMap{},
	}
	client := his.ProvideHISClient(cfg)
	_, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.ErrorIs(t, err, his.ErrHISNotConfigured)
}

// ❌ gender ไม่ถูกต้อง (ไม่ใช่ M หรือ F) → error
func TestSearchPatient_InvalidGender(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(his.HISPatient{PatientHN: "HN001", Gender: "X"})
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	_, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.Error(t, err)
}

// ❌ date_of_birth ผิดรูปแบบ (ไม่ใช่ YYYY-MM-DD) → error
func TestSearchPatient_InvalidDateOfBirth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(his.HISPatient{PatientHN: "HN001", Gender: "M", DateOfBirth: "20-05-1990"})
	}))
	defer srv.Close()

	client := his.ProvideHISClient(newHTTPCfg(srv.URL))
	_, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.Error(t, err)
}

// ✅ mock: Thai patient (hospital-a)
func TestMockClient_ThaiPatient(t *testing.T) {
	client := his.ProvideHISClient(config.AppConfig{HISMode: "mock"})

	got, err := client.SearchPatient(context.Background(), "hospital-a", "1234567890123")

	require.NoError(t, err)
	assert.Equal(t, "HNA001", got.PatientHN)
	assert.Equal(t, "1234567890123", got.NationalID)
	assert.Equal(t, "M", got.Gender)
	assert.Equal(t, "1990-05-20", got.DateOfBirth)
}

// ✅ mock: foreign patient (hospital-b) มีเฉพาะ passport_id
func TestMockClient_ForeignPatient(t *testing.T) {
	client := his.ProvideHISClient(config.AppConfig{HISMode: "mock"})

	got, err := client.SearchPatient(context.Background(), "hospital-b", "BB9876543")

	require.NoError(t, err)
	assert.Equal(t, "HNB002", got.PatientHN)
	assert.Equal(t, "BB9876543", got.PassportID)
	assert.Equal(t, "", got.NationalID)
	assert.Equal(t, "F", got.Gender)
}

// ❌ mock: patient ไม่พบ → ErrPatientNotFound
func TestMockClient_PatientNotFound(t *testing.T) {
	client := his.ProvideHISClient(config.AppConfig{HISMode: "mock"})
	_, err := client.SearchPatient(context.Background(), "hospital-a", "0000000000000")
	require.ErrorIs(t, err, his.ErrPatientNotFound)
}

// ❌ mock: hospital ไม่มีใน mock data → ErrHISNotConfigured
func TestMockClient_HospitalNotConfigured(t *testing.T) {
	client := his.ProvideHISClient(config.AppConfig{HISMode: "mock"})
	_, err := client.SearchPatient(context.Background(), "unknown-hospital", "1234567890123")
	require.ErrorIs(t, err, his.ErrHISNotConfigured)
}
