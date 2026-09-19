package entity

import "time"

type StaffHospital struct {
	StaffID    uint      `gorm:"primaryKey;not null"`
	HospitalID uint      `gorm:"primaryKey;not null"`
	CreatedAt  time.Time
}
