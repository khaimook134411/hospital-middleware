package staff_test

import (
	"bytes"
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
	staffhandler "github.com/khaimook/hospital-middleware/handler/staff"
	"github.com/khaimook/hospital-middleware/handler/middleware"
	mocksstaff "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/service/staff"
	"github.com/khaimook/hospital-middleware/util"
)

func init() { gin.SetMode(gin.TestMode) }

var testCfg = config.AppConfig{
	JWTSecret: "test-secret-key-that-is-32-chars!!",
	JWTTTL:    24 * time.Hour,
}

func adminToken() string {
	t, _ := util.GenerateToken(testCfg, 1, entity.RoleAdmin, nil, "")
	return t
}

func staffToken() string {
	id := uint(1)
	t, _ := util.GenerateToken(testCfg, 2, entity.RoleStaff, &id, "hospital-a")
	return t
}

func buildCreateEngine(svc *mocksstaff.MockStaffService) *gin.Engine {
	r := gin.New()
	h := staffhandler.NewHandler(svc, testCfg)
	r.POST("/staff/create",
		middleware.RequireAuth(testCfg),
		middleware.RequireRole(entity.RoleAdmin),
		h.Create,
	)
	return r
}

func postCreate(r *gin.Engine, body any, token string) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}

func assertErrCode(t *testing.T, body []byte, code string) {
	t.Helper()
	var resp struct {
		Error struct{ Code string `json:"code"` } `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, code, resp.Error.Code)
}

// ✅ admin สร้างสำเร็จ → 201, ไม่มี password_hash ใน response
func TestCreate_Success(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	now := time.Now()
	svc.EXPECT().Create(mock.Anything, "nurse01", "P@ssw0rd123", "hospital-a").
		Return(&entity.StaffResponse{ID: 2, Username: "nurse01", Hospital: "hospital-a", CreatedAt: now}, nil)

	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, adminToken())

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp struct {
		Data struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Hospital string `json:"hospital"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, uint(2), resp.Data.ID)
	assert.Equal(t, "nurse01", resp.Data.Username)
	assert.Equal(t, "hospital-a", resp.Data.Hospital)
	assert.NotContains(t, w.Body.String(), "password_hash")
}

// ✅ admin สร้าง staff ให้ hospital-b ได้ด้วย
func TestCreate_SuccessHospitalB(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Create(mock.Anything, "nurse02", "P@ssw0rd123", "hospital-b").
		Return(&entity.StaffResponse{ID: 3, Username: "nurse02", Hospital: "hospital-b"}, nil)

	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse02", "password": "P@ssw0rd123", "hospital": "hospital-b",
	}, adminToken())

	assert.Equal(t, http.StatusCreated, w.Code)
}

// ❌ ไม่มี token → 401
func TestCreate_NoToken(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertErrCode(t, w.Body.Bytes(), "UNAUTHORIZED")
}

// ❌ token ผิด → 401
func TestCreate_InvalidToken(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, "invalid.token.here")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ❌ ใช้ token ของ staff → 403
func TestCreate_StaffTokenForbidden(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, staffToken())
	assert.Equal(t, http.StatusForbidden, w.Code)
	assertErrCode(t, w.Body.Bytes(), "FORBIDDEN")
}

// ❌ username ซ้ำ → 409
func TestCreate_UsernameTaken(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Create(mock.Anything, "existing", "P@ssw0rd123", "hospital-a").
		Return(nil, util.ErrUsernameTaken())

	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "existing", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, adminToken())
	assert.Equal(t, http.StatusConflict, w.Code)
	assertErrCode(t, w.Body.Bytes(), "USERNAME_TAKEN")
}

// ❌ hospital ไม่มีอยู่ → 404
func TestCreate_HospitalNotFound(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	svc.EXPECT().Create(mock.Anything, "nurse01", "P@ssw0rd123", "no-such").
		Return(nil, util.ErrHospitalNotFound())

	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse01", "password": "P@ssw0rd123", "hospital": "no-such",
	}, adminToken())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrCode(t, w.Body.Bytes(), "HOSPITAL_NOT_FOUND")
}

// ❌ username สั้นเกิน (< 3) → 400
func TestCreate_UsernameTooShort(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "ab", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, adminToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

// ❌ username มีอักขระผิด (@, #) → 400
func TestCreate_UsernameInvalidChars(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse@01", "password": "P@ssw0rd123", "hospital": "hospital-a",
	}, adminToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

// ❌ password สั้นเกิน (< 8) → 400
func TestCreate_PasswordTooShort(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	w := postCreate(buildCreateEngine(svc), map[string]string{
		"username": "nurse01", "password": "short", "hospital": "hospital-a",
	}, adminToken())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

// ❌ field หาย → 400
func TestCreate_MissingFields(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	cases := []map[string]string{
		{"password": "P@ssw0rd123", "hospital": "hospital-a"},          // no username
		{"username": "nurse01", "hospital": "hospital-a"},              // no password
		{"username": "nurse01", "password": "P@ssw0rd123"},             // no hospital
	}
	for _, body := range cases {
		w := postCreate(buildCreateEngine(svc), body, adminToken())
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertErrCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
	}
}

// ❌ body ไม่ใช่ JSON → 400
func TestCreate_InvalidJSON(t *testing.T) {
	svc := mocksstaff.NewMockStaffService(t)
	r := buildCreateEngine(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
