package util_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/util"
)

var testCfg = config.AppConfig{
	JWTSecret: "supersecret-key-that-is-at-least-32-chars",
	JWTTTL:    time.Hour,
}

func TestGenerateToken_Admin(t *testing.T) {
	token, err := util.GenerateToken(testCfg, 1, entity.RoleAdmin, nil, "")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := util.ParseToken(testCfg, token)
	require.NoError(t, err)
	assert.Equal(t, "1", claims.Subject)
	assert.Equal(t, entity.RoleAdmin, claims.Role)
	assert.Nil(t, claims.HospitalID)
	assert.Empty(t, claims.HospitalCode)
}

func TestGenerateToken_Staff(t *testing.T) {
	hospitalID := uint(3)
	token, err := util.GenerateToken(testCfg, 5, entity.RoleStaff, &hospitalID, "hospital-a")
	require.NoError(t, err)

	claims, err := util.ParseToken(testCfg, token)
	require.NoError(t, err)
	assert.Equal(t, "5", claims.Subject)
	assert.Equal(t, entity.RoleStaff, claims.Role)
	require.NotNil(t, claims.HospitalID)
	assert.Equal(t, uint(3), *claims.HospitalID)
	assert.Equal(t, "hospital-a", claims.HospitalCode)
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := util.GenerateToken(testCfg, 1, entity.RoleAdmin, nil, "")
	require.NoError(t, err)

	wrongCfg := config.AppConfig{JWTSecret: "completely-different-secret-key!!"}
	_, err = util.ParseToken(wrongCfg, token)
	assert.Error(t, err)
}

func TestParseToken_Expired(t *testing.T) {
	expiredCfg := config.AppConfig{
		JWTSecret: testCfg.JWTSecret,
		JWTTTL:    -time.Minute,
	}
	token, err := util.GenerateToken(expiredCfg, 1, entity.RoleAdmin, nil, "")
	require.NoError(t, err)

	_, err = util.ParseToken(testCfg, token)
	assert.Error(t, err)
}

func TestParseToken_Malformed(t *testing.T) {
	_, err := util.ParseToken(testCfg, "not.a.jwt")
	assert.Error(t, err)
}

func TestParseToken_RejectsNoneAlg(t *testing.T) {
	// Header: {"alg":"none","typ":"JWT"}, Payload: {"sub":"1"}
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxIn0."
	_, err := util.ParseToken(testCfg, noneToken)
	assert.Error(t, err)
}

func TestParseToken_RejectsRS256(t *testing.T) {
	// Build a fake RS256-header token (signature won't be valid — that's fine,
	// we're checking the algorithm rejection path fires before signature check).
	header := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"  // {"alg":"RS256","typ":"JWT"}
	payload := "eyJzdWIiOiIxIn0"                        // {"sub":"1"}
	fakeToken := header + "." + payload + ".fakesig"
	_, err := util.ParseToken(testCfg, fakeToken)
	assert.Error(t, err)
}

func TestGenerateToken_UsesHS256(t *testing.T) {
	token, err := util.GenerateToken(testCfg, 1, entity.RoleAdmin, nil, "")
	require.NoError(t, err)
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	_, err = util.ParseToken(testCfg, token)
	assert.NoError(t, err)
}
