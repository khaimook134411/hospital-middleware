package testsuite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
	"github.com/khaimook/hospital-middleware/entity/migrator"
)

func NewPostgres(t testing.TB) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"docker.io/postgres:17-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("warning: terminate postgres container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpg.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	require.NoError(t, migrator.Migrate(db, config.AppConfig{}))

	return db
}

func TruncateAll(t testing.TB, db *gorm.DB) {
	t.Helper()
	err := db.Exec(
		"TRUNCATE TABLE patients, staff_hospitals, staffs, hospitals RESTART IDENTITY CASCADE",
	).Error
	require.NoError(t, err, "TruncateAll: truncate")

	hospitals := []entity.Hospital{
		{Code: "hospital-a", Name: "Hospital A"},
		{Code: "hospital-b", Name: "Hospital B"},
	}
	require.NoError(t, db.Create(&hospitals).Error, "TruncateAll: seed hospitals")
}
