package patient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/handler/middleware"
	patienthandler "github.com/khaimook/hospital-middleware/handler/patient"
	mockspacient "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/service/patient"
	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	"github.com/khaimook/hospital-middleware/util"
)

func init() { gin.SetMode(gin.TestMode) }

var testCfg = config.AppConfig{
	JWTSecret: "test-secret-key-that-is-32-chars!!",
	JWTTTL:    24 * time.Hour,
}

func staffToken() string {
	hID := uint(1)
	t, _ := util.GenerateToken(testCfg, 5, entity.RoleStaff, &hID, "hospital-a")
	return t
}

func adminToken() string {
	t, _ := util.GenerateToken(testCfg, 1, entity.RoleAdmin, nil, "")
	return t
}

func buildEngine(svc *mockspacient.MockPatientService) *gin.Engine {
	r := gin.New()
	h := patienthandler.NewHandler(svc)
	r.GET("/patient/search",
		middleware.RequireAuth(testCfg),
		middleware.RequireRole(entity.RoleStaff),
		h.Search,
	)
	return r
}

func doSearch(r *gin.Engine, query string, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/patient/search?"+query, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}

func assertErrCode(t *testing.T, body []byte, code string) {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, code, resp.Error.Code)
}

func TestSearch_Success(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	nid := "1234567890123"
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		NationalID: "1234567890123", Limit: 20, Offset: 0,
	}).Return([]entity.PatientResponse{{PatientHN: "HNA001", NationalID: &nid}}, nil)

	w := doSearch(buildEngine(svc), "national_id=1234567890123", staffToken())

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data []entity.PatientResponse `json:"data"`
		Meta struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
			Count  int `json:"count"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "HNA001", resp.Data[0].PatientHN)
	assert.Equal(t, 20, resp.Meta.Limit)
	assert.Equal(t, 0, resp.Meta.Offset)
	assert.Equal(t, 1, resp.Meta.Count)
}

func TestSearch_NoParams(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		Limit: 20, Offset: 0,
	}).Return([]entity.PatientResponse{}, nil)

	w := doSearch(buildEngine(svc), "", staffToken())

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data []entity.PatientResponse `json:"data"`
		Meta struct {
			Count int `json:"count"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	assert.Equal(t, 0, resp.Meta.Count)
}

func TestSearch_CustomLimitOffset(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		Limit: 10, Offset: 5,
	}).Return([]entity.PatientResponse{}, nil)

	w := doSearch(buildEngine(svc), "limit=10&offset=5", staffToken())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearch_EmptyResult(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		LastName: "Notfound", Limit: 20,
	}).Return([]entity.PatientResponse{}, nil)

	w := doSearch(buildEngine(svc), "last_name=Notfound", staffToken())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":[]`)
}

func TestSearch_NoToken(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertErrCode(t, w.Body.Bytes(), "UNAUTHORIZED")
}

func TestSearch_InvalidToken(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "", "invalid.token.here")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearch_AdminToken(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "", adminToken())
	assert.Equal(t, http.StatusForbidden, w.Code)
	assertErrCode(t, w.Body.Bytes(), "FORBIDDEN")
}

func TestSearch_NationalIDWrongLength(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "national_id=123456", staffToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

func TestSearch_NationalIDNotDigits(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "national_id=123456789012A", staffToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

func TestSearch_DateOfBirthInvalid(t *testing.T) {
	cases := []string{"20-05-1990", "1990/05/20", "not-a-date"}
	for _, dob := range cases {
		svc := mockspacient.NewMockPatientService(t)
		w := doSearch(buildEngine(svc), "date_of_birth="+dob, staffToken())
		assert.Equal(t, http.StatusBadRequest, w.Code, "dob=%s", dob)
		assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
	}
}

func TestSearch_EmailInvalid(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "email=notanemail", staffToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

func TestSearch_LimitTooHigh(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	w := doSearch(buildEngine(svc), "limit=101", staffToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

func TestSearch_LimitZeroUsesDefault(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		Limit: 20,
	}).Return([]entity.PatientResponse{}, nil)

	w := doSearch(buildEngine(svc), "limit=0", staffToken())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearch_HISUnavailable(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		NationalID: "1234567890123", Limit: 20,
	}).Return(nil, util.ErrHISUnavailable())

	w := doSearch(buildEngine(svc), "national_id=1234567890123", staffToken())
	assert.Equal(t, http.StatusBadGateway, w.Code)
	assertErrCode(t, w.Body.Bytes(), "HIS_UNAVAILABLE")
}

func TestSearch_SpecialCharsInName(t *testing.T) {
	svc := mockspacient.NewMockPatientService(t)
	svc.EXPECT().Search(mock.Anything, uint(1), "hospital-a", patientrepo.SearchParams{
		FirstName: "S%mch_y", Limit: 20,
	}).Return([]entity.PatientResponse{}, nil)

	w := doSearch(buildEngine(svc), "first_name=S%25mch_y", staffToken())
	assert.Equal(t, http.StatusOK, w.Code)
}
