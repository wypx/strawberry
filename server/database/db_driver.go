package database

import (
	"github.com/jinzhu/gorm"
	Model "tomato/models"
)

var DB *gorm.DB

func GetDB() *gorm.DB {
	return DB
}

func SetDB(db *gorm.DB)  {
	if DB == nil {
		DB = db
		db.AutoMigrate(&Model.User{})
	}
}