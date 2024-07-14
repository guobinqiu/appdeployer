package model

import (
	"time"
)

type User struct {
	ID        int        `gorm:"column:id;primary_key" json:"id"`
	Username  string     `gorm:"column:username;not null" json:"username"`
	Password  string     `gorm:"column:password;not null" json:"password"` //加密后的密码
	ClientID  int        `gorm:"column:client_id;not null" json:"client_id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"-"`
}
