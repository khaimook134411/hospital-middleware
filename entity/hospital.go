package entity

import "time"

type Hospital struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Code      string    `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
