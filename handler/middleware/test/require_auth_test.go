package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/handler/middleware"
	"github.com/khaimook/hospital-middleware/util"
)

func init() { gin.SetMode(gin.TestMode) }

var authCfg = config.AppConfig{
	JWTSecret: "test-secret-key-that-is-32-chars!!",
	JWTTTL:    time.Hour,
}

func buildAuthEngine() (*gin.Engine, *string) {
	capturedRole := new(string)
	r := gin.New()
	r.GET("/", middleware.RequireAuth(authCfg), func(c *gin.Context) {
		claims := util.GetClaims(c)
		*capturedRole = claims.Role
		c.Status(http.StatusOK)
	})
	return r, capturedRole
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	r, _ := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assertErrorCode(t, w.Body.Bytes(), "UNAUTHORIZED")
}

func TestRequireAuth_WrongFormat(t *testing.T) {
	r, _ := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	r, _ := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	expiredCfg := config.AppConfig{JWTSecret: authCfg.JWTSecret, JWTTTL: -time.Minute}
	token, err := util.GenerateToken(expiredCfg, 1, entity.RoleAdmin, nil, "")
	require.NoError(t, err)

	r, _ := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_NoneAlg(t *testing.T) {
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxIn0."
	r, _ := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+noneToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ValidAdminToken(t *testing.T) {
	token, err := util.GenerateToken(authCfg, 1, entity.RoleAdmin, nil, "")
	require.NoError(t, err)

	r, capturedRole := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, entity.RoleAdmin, *capturedRole)
}

func TestRequireAuth_ValidStaffToken(t *testing.T) {
	hospitalID := uint(1)
	token, err := util.GenerateToken(authCfg, 5, entity.RoleStaff, &hospitalID, "hospital-a")
	require.NoError(t, err)

	r, capturedRole := buildAuthEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, entity.RoleStaff, *capturedRole)
}

func assertErrorCode(t *testing.T, body []byte, code string) {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, code, resp.Error.Code)
}
