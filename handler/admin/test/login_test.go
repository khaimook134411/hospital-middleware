package admin_test

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
	adminhandler "github.com/khaimook/hospital-middleware/handler/admin"
	mocksadmin "github.com/khaimook/hospital-middleware/mocks/github.com/khaimook/hospital-middleware/service/admin"
	"github.com/khaimook/hospital-middleware/util"
)

func init() { gin.SetMode(gin.TestMode) }

var handlerCfg = config.AppConfig{
	JWTSecret: "test-secret-key-that-is-32-chars!!",
	JWTTTL:    24 * time.Hour,
}

func buildEngine(svc *mocksadmin.MockAdminService) *gin.Engine {
	r := gin.New()
	h := adminhandler.NewHandler(svc, handlerCfg)
	r.POST("/admin/login", h.Login)
	return r
}

func postJSON(r *gin.Engine, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestAdminLogin_Success(t *testing.T) {
	svc := mocksadmin.NewMockAdminService(t)
	svc.EXPECT().Login(anyCtx(), "admin", "pass123").Return("fake.jwt.token", nil)

	w := postJSON(buildEngine(svc), map[string]string{"username": "admin", "password": "pass123"})

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

func TestAdminLogin_InvalidCredentials(t *testing.T) {
	svc := mocksadmin.NewMockAdminService(t)
	svc.EXPECT().Login(anyCtx(), "admin", "wrong").Return("", util.ErrInvalidCredentials())

	w := postJSON(buildEngine(svc), map[string]string{"username": "admin", "password": "wrong"})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertCode(t, w.Body.Bytes(), "INVALID_CREDENTIALS")
}

func TestAdminLogin_MissingUsername(t *testing.T) {
	svc := mocksadmin.NewMockAdminService(t)
	w := postJSON(buildEngine(svc), map[string]string{"password": "pass"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

func TestAdminLogin_MissingPassword(t *testing.T) {
	svc := mocksadmin.NewMockAdminService(t)
	w := postJSON(buildEngine(svc), map[string]string{"username": "admin"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertCode(t, w.Body.Bytes(), "VALIDATION_ERROR")
}

func TestAdminLogin_InvalidJSON(t *testing.T) {
	svc := mocksadmin.NewMockAdminService(t)
	r := buildEngine(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminLogin_InternalError(t *testing.T) {
	svc := mocksadmin.NewMockAdminService(t)
	svc.EXPECT().Login(anyCtx(), "admin", "pass").Return("", util.ErrInternal())

	w := postJSON(buildEngine(svc), map[string]string{"username": "admin", "password": "pass"})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assertCode(t, w.Body.Bytes(), "INTERNAL_ERROR")
}

// helpers

func anyCtx() any {
	return mock.Anything
}

func assertCode(t *testing.T, body []byte, code string) {
	t.Helper()
	var resp struct {
		Error struct{ Code string `json:"code"` } `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, code, resp.Error.Code)
}
