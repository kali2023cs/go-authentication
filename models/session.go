package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null;index"`
	Token     string         `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time      `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
