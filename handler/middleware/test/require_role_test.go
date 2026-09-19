package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/handler/middleware"
	"github.com/khaimook/hospital-middleware/util"
)

func buildRoleEngine(requiredRole string, claimsRole string) *gin.Engine {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.Set("claims", &util.Claims{Role: claimsRole})
		c.Next()
	}, middleware.RequireRole(requiredRole), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRequireRole_Matching(t *testing.T) {
	cases := []struct {
		required string
		actual   string
	}{
		{entity.RoleAdmin, entity.RoleAdmin},
		{entity.RoleStaff, entity.RoleStaff},
	}
	for _, tc := range cases {
		r := buildRoleEngine(tc.required, tc.actual)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "role %s should pass", tc.required)
	}
}

func TestRequireRole_Mismatch(t *testing.T) {
	cases := []struct {
		required string
		actual   string
	}{
		{entity.RoleAdmin, entity.RoleStaff},
		{entity.RoleStaff, entity.RoleAdmin},
	}
	for _, tc := range cases {
		r := buildRoleEngine(tc.required, tc.actual)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code, "role %s should be forbidden when actual is %s", tc.required, tc.actual)
		assertErrorCode(t, w.Body.Bytes(), "FORBIDDEN")
	}
}
