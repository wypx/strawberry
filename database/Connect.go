package database

import (
	"tomato/config"
	"tomato/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	conf := config.Get()
	db, err := gorm.Open("mysql", conf.DSN)

	if err == nil {
		db.DB().SetMaxIdleConns(conf.MaxIdleConn)
		DB = db
		db.AutoMigrate(&models.AdminUser{}, &models.User{}, &models.Role{}, &models.Urls{})
		db.Model(&models.User{}).AddForeignKey("role_id", "roles(id)", "no action", "no action")
		return db, err
	}
	return nil, err
}
