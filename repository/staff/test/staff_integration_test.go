//go:build integration

package staff_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
	staffrepo "github.com/khaimook/hospital-middleware/repository/staff"
	"github.com/khaimook/hospital-middleware/testsuite"
	"github.com/khaimook/hospital-middleware/util"
)

func TestStaffRepository(t *testing.T) {
	db := testsuite.NewPostgres(t)
	repo := staffrepo.ProvideStaffRepository(db)
	ctx := context.Background()

	newStaff := func(username string) *entity.Staff {
		hash, _ := util.HashPassword("P@ssw0rd123")
		return &entity.Staff{Username: username, PasswordHash: hash, Role: entity.RoleStaff}
	}

	t.Run("CreateWithHospital — creates staff and mapping", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s := newStaff("nurse01")
		err := repo.CreateWithHospital(ctx, s, 1) // hospital-a = ID 1

		require.NoError(t, err)
		assert.NotZero(t, s.ID)

		var count int64
		db.Table("staff_hospitals").Where("staff_id = ? AND hospital_id = ?", s.ID, 1).Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("CreateWithHospital — same staff can map to multiple hospitals", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s := newStaff("nurse_multi")
		require.NoError(t, repo.CreateWithHospital(ctx, s, 1))

		db.Exec("INSERT INTO staff_hospitals (staff_id, hospital_id) VALUES (?, ?)", s.ID, 2)

		var count int64
		db.Table("staff_hospitals").Where("staff_id = ?", s.ID).Count(&count)
		assert.Equal(t, int64(2), count)
	})

	t.Run("CreateWithHospital — duplicate username returns pgconn 23505", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s1 := newStaff("duplicate_user")
		require.NoError(t, repo.CreateWithHospital(ctx, s1, 1))

		s2 := newStaff("duplicate_user")
		err := repo.CreateWithHospital(ctx, s2, 2)

		require.Error(t, err)
		var pgErr *pgconn.PgError
		assert.ErrorAs(t, err, &pgErr)
		assert.Equal(t, "23505", pgErr.Code)
	})

	t.Run("CreateWithHospital — invalid hospital_id rolls back staff creation", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s := newStaff("nurse_rollback")
		err := repo.CreateWithHospital(ctx, s, 99999)

		require.Error(t, err)

		_, findErr := repo.FindByUsername(ctx, "nurse_rollback")
		require.ErrorIs(t, findErr, gorm.ErrRecordNotFound)
	})

	t.Run("FindByUsername — existing user", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s := newStaff("nurse_find")
		require.NoError(t, repo.CreateWithHospital(ctx, s, 1))

		found, err := repo.FindByUsername(ctx, "nurse_find")

		require.NoError(t, err)
		assert.Equal(t, "nurse_find", found.Username)
		assert.Equal(t, entity.RoleStaff, found.Role)
		assert.NotEmpty(t, found.PasswordHash)
	})

	t.Run("FindByUsername — not found returns ErrRecordNotFound", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		_, err := repo.FindByUsername(ctx, "nobody")

		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("HasHospital — staff has mapping returns true", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s := newStaff("nurse_hasmapping")
		require.NoError(t, repo.CreateWithHospital(ctx, s, 1))

		ok, err := repo.HasHospital(ctx, s.ID, 1)

		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("HasHospital — staff has no mapping returns false", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		s := newStaff("nurse_nomapping")
		require.NoError(t, repo.CreateWithHospital(ctx, s, 1))

		ok, err := repo.HasHospital(ctx, s.ID, 2) // hospital-b

		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("HasHospital — non-existent staff returns false", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		ok, err := repo.HasHospital(ctx, 99999, 1)

		require.NoError(t, err)
		assert.False(t, ok)
	})
}
