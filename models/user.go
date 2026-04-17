package models

type User struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `json:"name"`
	Email      string `json:"email" gorm:"unique"`
	Phone      string `json:"phone" gorm:"unique"`
	Password   string `json:"password"`
	GoogleID   string `json:"google_id" gorm:"uniqueIndex"`
	IsVerified bool   `json:"is_verified" gorm:"default:false"`
}
