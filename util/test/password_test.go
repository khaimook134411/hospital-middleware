package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khaimook/hospital-middleware/util"
)

func TestHashPassword_ProducesBcrypt(t *testing.T) {
	hash, err := util.HashPassword("P@ssw0rd123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "P@ssw0rd123", hash)
}

func TestCheckPassword_Correct(t *testing.T) {
	hash, err := util.HashPassword("P@ssw0rd123")
	require.NoError(t, err)
	assert.True(t, util.CheckPassword(hash, "P@ssw0rd123"))
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, err := util.HashPassword("P@ssw0rd123")
	require.NoError(t, err)
	assert.False(t, util.CheckPassword(hash, "wrongpassword"))
}

func TestCheckPassword_DifferentHashes(t *testing.T) {
	hash1, _ := util.HashPassword("samepassword")
	hash2, _ := util.HashPassword("samepassword")
	assert.NotEqual(t, hash1, hash2)
	assert.True(t, util.CheckPassword(hash1, "samepassword"))
	assert.True(t, util.CheckPassword(hash2, "samepassword"))
}
