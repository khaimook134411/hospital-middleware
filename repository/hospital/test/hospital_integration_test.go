//go:build integration

package hospital_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	hospitalrepo "github.com/khaimook/hospital-middleware/repository/hospital"
	"github.com/khaimook/hospital-middleware/testsuite"
)

func TestHospitalRepository(t *testing.T) {
	db := testsuite.NewPostgres(t)
	repo := hospitalrepo.ProvideHospitalRepository(db)
	ctx := context.Background()

	t.Run("FindByCode — hospital-a exists", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		h, err := repo.FindByCode(ctx, "hospital-a")

		require.NoError(t, err)
		assert.Equal(t, "hospital-a", h.Code)
		assert.Equal(t, "Hospital A", h.Name)
		assert.NotZero(t, h.ID)
	})

	t.Run("FindByCode — hospital-b exists", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		h, err := repo.FindByCode(ctx, "hospital-b")

		require.NoError(t, err)
		assert.Equal(t, "hospital-b", h.Code)
		assert.Equal(t, "Hospital B", h.Name)
	})

	t.Run("FindByCode — not found returns ErrRecordNotFound", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		_, err := repo.FindByCode(ctx, "hospital-does-not-exist")

		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("FindByCode — hospital IDs are stable after truncate+reseed", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		ha, err := repo.FindByCode(ctx, "hospital-a")
		require.NoError(t, err)
		hb, err := repo.FindByCode(ctx, "hospital-b")
		require.NoError(t, err)

		assert.Equal(t, uint(1), ha.ID)
		assert.Equal(t, uint(2), hb.ID)
	})
}
