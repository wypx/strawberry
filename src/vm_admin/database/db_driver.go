package database

import (
	Model "vm_manager/vm_admin/models"

	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func GetDB() *gorm.DB {
	return DB
}

func SetDB(db *gorm.DB) {
	if DB == nil {
		DB = db
		db.AutoMigrate(&Model.User{})
	}
}
