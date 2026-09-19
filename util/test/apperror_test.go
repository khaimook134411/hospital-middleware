package util_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/khaimook/hospital-middleware/util"
)

func TestAppErrorCodes(t *testing.T) {
	cases := []struct {
		err    *util.AppError
		status int
		code   string
	}{
		{util.ErrValidation("bad input"), http.StatusBadRequest, "VALIDATION_ERROR"},
		{util.ErrUnauthorized(), http.StatusUnauthorized, "UNAUTHORIZED"},
		{util.ErrInvalidCredentials(), http.StatusUnauthorized, "INVALID_CREDENTIALS"},
		{util.ErrForbidden(), http.StatusForbidden, "FORBIDDEN"},
		{util.ErrHospitalNotFound(), http.StatusNotFound, "HOSPITAL_NOT_FOUND"},
		{util.ErrUsernameTaken(), http.StatusConflict, "USERNAME_TAKEN"},
		{util.ErrHISUnavailable(), http.StatusBadGateway, "HIS_UNAVAILABLE"},
		{util.ErrInternal(), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.status, tc.err.Status, tc.code)
		assert.Equal(t, tc.code, tc.err.Code)
		assert.NotEmpty(t, tc.err.Message)
		assert.NotEmpty(t, tc.err.Error())
	}
}

func TestErrValidation_CustomMessage(t *testing.T) {
	err := util.ErrValidation("username is required")
	assert.Equal(t, "username is required", err.Message)
}
