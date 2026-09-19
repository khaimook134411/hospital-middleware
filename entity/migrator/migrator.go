package migrator

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/khaimook/hospital-middleware/di/config"
	"github.com/khaimook/hospital-middleware/entity"
)

func Migrate(db *gorm.DB, cfg config.AppConfig) error {
	if err := db.AutoMigrate(
		&entity.Hospital{},
		&entity.Staff{},
		&entity.StaffHospital{},
		&entity.Patient{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	if err := runRawSQL(db); err != nil {
		return err
	}

	if err := seedHospitals(db); err != nil {
		return err
	}

	return seedAdmin(db, cfg)
}

func runRawSQL(db *gorm.DB) error {
	stmts := []string{
		// CHECK constraints (no IF NOT EXISTS in Postgres; use DO block)
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_staffs_role') THEN
				ALTER TABLE staffs ADD CONSTRAINT chk_staffs_role CHECK (role IN ('admin','staff'));
			END IF;
		END $$`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_patients_gender') THEN
				ALTER TABLE patients ADD CONSTRAINT chk_patients_gender CHECK (gender IN ('M','F'));
			END IF;
		END $$`,

		// FK on staff_hospitals with ON DELETE CASCADE
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_staff_hospitals_staff') THEN
				ALTER TABLE staff_hospitals
					ADD CONSTRAINT fk_staff_hospitals_staff
					FOREIGN KEY (staff_id) REFERENCES staffs(id) ON DELETE CASCADE;
			END IF;
		END $$`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_staff_hospitals_hospital') THEN
				ALTER TABLE staff_hospitals
					ADD CONSTRAINT fk_staff_hospitals_hospital
					FOREIGN KEY (hospital_id) REFERENCES hospitals(id) ON DELETE CASCADE;
			END IF;
		END $$`,

		// FK on patients
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_patients_hospital') THEN
				ALTER TABLE patients
					ADD CONSTRAINT fk_patients_hospital
					FOREIGN KEY (hospital_id) REFERENCES hospitals(id);
			END IF;
		END $$`,

		// Unique indexes on patients
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_hn ON patients (hospital_id, patient_hn)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_national_id ON patients (hospital_id, national_id) WHERE national_id IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_passport_id ON patients (hospital_id, passport_id) WHERE passport_id IS NOT NULL`,

		// Compound search indexes on patients
		`CREATE INDEX IF NOT EXISTS idx_patient_hospital_last_name_en ON patients (hospital_id, last_name_en)`,
		`CREATE INDEX IF NOT EXISTS idx_patient_hospital_last_name_th ON patients (hospital_id, last_name_th)`,
		`CREATE INDEX IF NOT EXISTS idx_patient_hospital_phone ON patients (hospital_id, phone_number)`,
		`CREATE INDEX IF NOT EXISTS idx_patient_hospital_email ON patients (hospital_id, email)`,

		// Index on staff_hospitals(hospital_id) for reverse lookup
		`CREATE INDEX IF NOT EXISTS idx_staff_hospitals_hospital_id ON staff_hospitals (hospital_id)`,
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("raw SQL: %w", err)
		}
	}
	return nil
}

func seedHospitals(db *gorm.DB) error {
	hospitals := []entity.Hospital{
		{Code: "hospital-a", Name: "Hospital A"},
		{Code: "hospital-b", Name: "Hospital B"},
	}
	for i := range hospitals {
		h := &hospitals[i]
		if err := db.Where(entity.Hospital{Code: h.Code}).FirstOrCreate(h).Error; err != nil {
			return fmt.Errorf("seed hospital %s: %w", h.Code, err)
		}
	}
	return nil
}

func seedAdmin(db *gorm.DB, cfg config.AppConfig) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Println("warning: ADMIN_USERNAME or ADMIN_PASSWORD not set, skipping admin seed")
		return nil
	}

	var existing entity.Staff
	err := db.Where("username = ?", cfg.AdminUsername).First(&existing).Error
	if err == nil {
		return nil // already exists — do not overwrite
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check admin: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := entity.Staff{
		Username:     cfg.AdminUsername,
		PasswordHash: string(hash),
		Role:         entity.RoleAdmin,
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	log.Printf("admin %q created", cfg.AdminUsername)
	return nil
}
