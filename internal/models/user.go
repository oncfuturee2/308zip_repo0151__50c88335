package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleDistributor Role = "distributor"
	RoleFinance     Role = "finance"
	RoleAdmin       Role = "admin"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Role      Role           `json:"role" gorm:"size:20;not null;default:'distributor'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
