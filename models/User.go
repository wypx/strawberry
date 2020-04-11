package models

import (
	"time"
)

type Role struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"type:varchar(32)"`
	Permission []Urls `gorm:"many2many:role_urls"`
}

type User struct {
	ID        uint   `gorm:"primary_key"`
	UserName  string `gorm:"type:varchar(64);unique;not null"`
	PassWord  string `gorm:"type:varchar(64);not null;default:''"`
	Email     string `gorm:"type:varchar(128);not null;default:''"`
	Address   string `gorm:"type:varchar(64);not null;default:''"`
	Phone     string `gorm:"type:varchar(64);not null;default:''"`
	WebSite   string `gorm:"type:varchar(64);not null;default:''"`
	NickName  string `gorm:"type:varchar(64);not null;default:''"`
	Sex       string `gorm:"type:varchar(16);"`
	Active    bool   `gorm:"not null;default:'1'"`
	SuperUser bool   `gorm:"not null;default:'1'"`
	Role      Role   `gorm:"foreignkey:RoleID;association_foreignkey:ID"`
	RoleID    uint   `json:"role_id"`
	Created  time.Time  `json:"created"`
	Updated  time.Time  `json:"updated"`
}