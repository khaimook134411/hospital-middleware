//go:build integration

package patient_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/entity"
	patientrepo "github.com/khaimook/hospital-middleware/repository/patient"
	"github.com/khaimook/hospital-middleware/testsuite"
)

func TestPatientRepository(t *testing.T) {
	db := testsuite.NewPostgres(t)
	repo := patientrepo.ProvidePatientRepository(db)
	ctx := context.Background()

	strPtr := func(s string) *string { return &s }

	dateOf := func(s string) *entity.Date {
		tt, _ := time.Parse("2006-01-02", s)
		d := entity.Date{Time: tt}
		return &d
	}

	seed := func(t *testing.T, p entity.Patient) entity.Patient {
		t.Helper()
		require.NoError(t, db.Create(&p).Error)
		return p
	}

	t.Run("Search — by national_id (exact match)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", NationalID: strPtr("1234567890123")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", NationalID: strPtr("9999999999999")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{NationalID: "1234567890123", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — by passport_id (exact match)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA003", PassportID: strPtr("AA1234567")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA004", PassportID: strPtr("BB9999999")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{PassportID: "AA1234567", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA003", results[0].PatientHN)
	})

	t.Run("Search — by first_name_th (ILIKE partial)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", FirstNameTH: strPtr("สมชาย")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", FirstNameTH: strPtr("วิชัย")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{FirstName: "สมช", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — by first_name_en (ILIKE, case insensitive)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", FirstNameEN: strPtr("Somchai")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", FirstNameEN: strPtr("Wichai")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{FirstName: "somchai", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — by last_name_en (ILIKE partial)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", LastNameEN: strPtr("Jaidee")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", LastNameEN: strPtr("Smith")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{LastName: "aide", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — by email (case insensitive LOWER)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", Email: strPtr("Somchai@Example.COM")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", Email: strPtr("other@test.com")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{Email: "somchai@example.com", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — by date_of_birth (exact)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", DateOfBirth: dateOf("1990-05-20")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", DateOfBirth: dateOf("2000-01-15")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{DateOfBirth: "1990-05-20", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — by phone_number (exact)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", PhoneNumber: strPtr("0812345678")})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002", PhoneNumber: strPtr("0899999999")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{PhoneNumber: "0812345678", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — multi-param AND (name AND DOB)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{
			HospitalID: 1, PatientHN: "HNA001",
			LastNameEN:  strPtr("Jaidee"),
			DateOfBirth: dateOf("1990-05-20"),
		})
		seed(t, entity.Patient{
			HospitalID: 1, PatientHN: "HNA002",
			LastNameEN:  strPtr("Jaidee"),
			DateOfBirth: dateOf("2000-01-01"), // different DOB
		})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{
			LastName: "Jaidee", DateOfBirth: "1990-05-20", Limit: 20,
		})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — no params returns all patients for that hospital", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001"})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA002"})
		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA003"})
		seed(t, entity.Patient{HospitalID: 2, PatientHN: "HNB001"}) // different hospital

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{Limit: 20})

		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("Search — hospital isolation (hospital-a cannot see hospital-b patients)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		nid := "1234567890123"
		seed(t, entity.Patient{HospitalID: 2, PatientHN: "HNB001", NationalID: strPtr(nid)})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{NationalID: nid, Limit: 20})

		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Search — % in first_name is escaped (not treated as wildcard)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", FirstNameEN: strPtr("Smith")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{FirstName: "Sm%", Limit: 20})

		require.NoError(t, err)
		assert.Empty(t, results, "% in search must not act as wildcard")
	})

	t.Run("Search — _ in first_name is escaped (not treated as single-char wildcard)", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", FirstNameEN: strPtr("Smith")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{FirstName: "Smi_h", Limit: 20})

		require.NoError(t, err)
		assert.Empty(t, results, "_ in search must not act as wildcard")
	})

	t.Run("Search — partial match works without special chars", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", FirstNameEN: strPtr("Smith")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{FirstName: "mith", Limit: 20})

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "HNA001", results[0].PatientHN)
	})

	t.Run("Search — limit and offset pagination", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		for i := 1; i <= 5; i++ {
			seed(t, entity.Patient{
				HospitalID: 1,
				PatientHN:  "HNA" + string(rune('0'+i)),
			})
		}

		page1, err := repo.Search(ctx, 1, patientrepo.SearchParams{Limit: 2, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		page2, err := repo.Search(ctx, 1, patientrepo.SearchParams{Limit: 2, Offset: 2})
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		page3, err := repo.Search(ctx, 1, patientrepo.SearchParams{Limit: 2, Offset: 4})
		require.NoError(t, err)
		assert.Len(t, page3, 1)

		assert.NotEqual(t, page1[0].PatientHN, page2[0].PatientHN)
	})

	t.Run("Search — empty result when no match", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		seed(t, entity.Patient{HospitalID: 1, PatientHN: "HNA001", NationalID: strPtr("1234567890123")})

		results, err := repo.Search(ctx, 1, patientrepo.SearchParams{NationalID: "0000000000000", Limit: 20})

		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Upsert — new patient is created", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		p := &entity.Patient{
			HospitalID:  1,
			PatientHN:   "HNA001",
			NationalID:  strPtr("1234567890123"),
			FirstNameEN: strPtr("Somchai"),
			LastNameEN:  strPtr("Jaidee"),
			Gender:      strPtr("M"),
			DateOfBirth: dateOf("1990-05-20"),
		}

		err := repo.Upsert(ctx, p)
		require.NoError(t, err)

		var saved entity.Patient
		require.NoError(t, db.First(&saved, "hospital_id = ? AND patient_hn = ?", 1, "HNA001").Error)
		assert.Equal(t, "Somchai", *saved.FirstNameEN)
		assert.Equal(t, "1234567890123", *saved.NationalID)
		assert.Equal(t, "M", *saved.Gender)
	})

	t.Run("Upsert — existing patient_hn updates all mutable fields", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		initial := &entity.Patient{
			HospitalID:  1,
			PatientHN:   "HNA001",
			NationalID:  strPtr("1234567890123"),
			FirstNameEN: strPtr("OldName"),
			Gender:      strPtr("M"),
		}
		require.NoError(t, repo.Upsert(ctx, initial))

		updated := &entity.Patient{
			HospitalID:  1,
			PatientHN:   "HNA001",
			NationalID:  strPtr("1234567890123"),
			FirstNameEN: strPtr("NewName"),
			LastNameEN:  strPtr("AddedLastName"),
			Gender:      strPtr("M"),
		}
		require.NoError(t, repo.Upsert(ctx, updated))

		var saved entity.Patient
		require.NoError(t, db.First(&saved, "hospital_id = ? AND patient_hn = ?", 1, "HNA001").Error)
		assert.Equal(t, "NewName", *saved.FirstNameEN)
		assert.Equal(t, "AddedLastName", *saved.LastNameEN)
	})

	t.Run("Upsert — same national_id in same hospital with different HN violates unique constraint", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		require.NoError(t, repo.Upsert(ctx, &entity.Patient{
			HospitalID: 1, PatientHN: "HNA001", NationalID: strPtr("1234567890123"),
		}))

		err := repo.Upsert(ctx, &entity.Patient{
			HospitalID: 1, PatientHN: "HNA002", NationalID: strPtr("1234567890123"),
		})

		require.Error(t, err, "duplicate national_id in same hospital must error")
		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr)
		assert.Equal(t, "23505", pgErr.Code)
	})

	t.Run("Upsert — same national_id in different hospital is allowed", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		require.NoError(t, repo.Upsert(ctx, &entity.Patient{
			HospitalID: 1, PatientHN: "HNA001", NationalID: strPtr("1234567890123"),
		}))
		require.NoError(t, repo.Upsert(ctx, &entity.Patient{
			HospitalID: 2, PatientHN: "HNB001", NationalID: strPtr("1234567890123"),
		}))

		var count int64
		db.Model(&entity.Patient{}).Where("national_id = ?", "1234567890123").Count(&count)
		assert.Equal(t, int64(2), count)
	})

	t.Run("Upsert — gender constraint rejects invalid value", func(t *testing.T) {
		testsuite.TruncateAll(t, db)

		err := repo.Upsert(ctx, &entity.Patient{
			HospitalID: 1, PatientHN: "HNA001", Gender: strPtr("X"),
		})

		require.Error(t, err)
		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr)
		assert.Equal(t, "23514", pgErr.Code)
	})
}

var _ = gorm.ErrRecordNotFound
