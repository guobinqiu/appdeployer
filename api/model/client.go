package model

import (
	"time"
)

type Client struct {
	ID           int        `gorm:"column:id;primary_key" json:"id"`
	ClientKey    string     `gorm:"column:client_key;not null" json:"clientKey"`
	ClientSecret string     `gorm:"column:client_secret;not null" json:"clientSecret"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `gorm:"column:deleted_at" json:"-"`
}
