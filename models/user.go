package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"not null" json:"username"`
	Email     string    `gorm:"unique not null" json:"email"`
	Fid       string    `gorm:"unique not null" json:"fid"`
	Role      string    `gorm:"not null" json:"role"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
