package models

import "time"

type OTP struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}
