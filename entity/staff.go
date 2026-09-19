package entity

import "time"

const (
	RoleAdmin = "admin"
	RoleStaff = "staff"
)

type Staff struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StaffResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Hospital  string    `json:"hospital"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Staff) ToResponse(hospitalCode string) StaffResponse {
	return StaffResponse{
		ID:        s.ID,
		Username:  s.Username,
		Hospital:  hospitalCode,
		CreatedAt: s.CreatedAt,
	}
}
