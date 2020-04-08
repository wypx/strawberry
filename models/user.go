package models

type User struct {
	ID        uint   `gorm:"primary_key"`
	Username  string `gorm:"type:varchar(64);unique;not null"`
	Password  string `gorm:"type:varchar(64);not null;default:''"`
	Email     string `gorm:"type:varchar(128);not null;default:''"`
	Nickname  string `gorm:"type:varchar(64);not null;default:''"`
	Sex       string `gorm:"type:varchar(16);"`
	Active    bool   `gorm:"not null;default:'1'"`
	Superuser bool   `gorm:"not null;default:'1'"`
	Role      Role   `gorm:"foreignkey:RoleID;association_foreignkey:ID"`
	RoleID    uint   `json:"role_id"`
}

type Role struct {
	ID         uint   `gorm:"primary_key"`
	Name       string `gorm:"type:varchar(32)"`
	Permission []Urls `gorm:"many2many:role_urls"`
}
