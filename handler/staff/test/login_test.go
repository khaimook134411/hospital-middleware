package staff_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	staffhandler "github.com/khaimook/hospital-middleware/handler/staff"
	mocksstaff "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/service/staff"
	"github.com/khaimook/hospital-middleware/util"
)

func buildLoginEngine(svc *mocksstaff.MockStaffService) *gin.Engine {
	r := gin.New()
	h := staffhandler.NewHandler(svc, testCfg)
	r.POST("/staff/login", h.Login)
	return r
}

func postLogin(r *gin.Engine, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// ✅ login สำเร็จ → 200, token, token_type, expires_in
func TestLogin_Success(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Login(mock.Anything, "nurse01", "P@ssw0rd123", "hospital-a").
		Return("fake.jwt.token", nil)

	w := postLogin(buildLoginEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "hospital-a",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int64  `json:"expires_in"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "fake.jwt.token", resp.Data.AccessToken)
	assert.Equal(t, "Bearer", resp.Data.TokenType)
	assert.Equal(t, int64(86400), resp.Data.ExpiresIn)
}

// ✅ staff ที่มี 2 โรงพยาบาล ได้ token ที่ hospital_id ต่างกัน
func TestLogin_TwoHospitalsGetDifferentTokens(t *testing.T) {
	for _, hospital := range []string{"hospital-a", "hospital-b"} {
		t.Run("login at "+hospital, func(t *testing.T) {
			svc := mocksstaff.NewMockStaffService(t)
			svc.EXPECT().Login(mock.Anything, "nurse01", "P@ssw0rd123", hospital).
				Return("token-for-"+hospital, nil)

			w := postLogin(buildLoginEngine(svc), map[string]string{
				"username": "nurse01", "password": "P@ssw0rd123", "hospital": hospital,
			})
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "token-for-"+hospital)
		})
	}
}

// ❌ password ผิด → 401
func TestLogin_WrongPassword(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Login(mock.Anything, "nurse01", "wrong", "hospital-a").
		Return("", util.ErrInvalidCredentials())

	w := postLogin(buildLoginEngine(svc), map[string]string{
		"username": "nurse01", "password": "wrong", "hospital": "hospital-a",
	})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertErrCode(t, w.Body.Bytes(), "INVALID_CREDENTIALS")
}

// ❌ username ไม่มี → 401
func TestLogin_UsernameNotFound(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Login(mock.Anything, "nobody", "P@ssw0rd123", "hospital-a").
		Return("", util.ErrInvalidCredentials())

	w := postLogin(buildLoginEngine(svc), map[string]string{
		"username": "nobody", "password": "P@ssw0rd123", "hospital": "hospital-a",
	})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ❌ hospital ที่ staff ไม่มี mapping → 401
func TestLogin_NoHospitalMapping(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Login(mock.Anything, "nurse01", "P@ssw0rd123", "hospital-b").
		Return("", util.ErrInvalidCredentials())

	w := postLogin(buildLoginEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "hospital-b",
	})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ❌ ใช้บัญชี admin login ที่ /staff/login → 401
func TestLogin_AdminRejected(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Login(mock.Anything, "admin", "adminpass", "hospital-a").
		Return("", util.ErrInvalidCredentials())

	w := postLogin(buildLoginEngine(svc), map[string]string{
		"username": "admin", "password": "adminpass", "hospital": "hospital-a",
	})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ❌ field หาย → 400
func TestLogin_MissingFields(t *testing.T) {
	cases := []map[string]string{
		{"password": "P@ssw0rd123", "hospital": "hospital-a"}, // no username
		{"username": "nurse01", "hospital": "hospital-a"},     // no password
		{"username": "nurse01", "password": "P@ssw0rd123"},    // no hospital
	}
	for _, body := range cases {
		svc := mocksstaff.NewMockStaffService(t)
		w := postLogin(buildLoginEngine(svc), body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
	}
}
