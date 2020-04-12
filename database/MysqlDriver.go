package database

import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/jinzhu/gorm"
    "tomato/config"
	"tomato/models"
)

var DbMysql *gorm.DB

func InitDB() (*gorm.DB, error) {
	conf := config.Get()
	db, err := gorm.Open("mysql", conf.DSN)

	if err == nil {
		db.DB().SetMaxIdleConns(conf.MaxIdleConn)
		DbMysql = db
		db.AutoMigrate(&models.AdminUser{}, &models.User{}, &models.Role{}, &models.Urls{})
		db.Model(&models.User{}).AddForeignKey("role_id", "roles(id)", "no action", "no action")
		return db, err
	}
	return nil, err
}

func init(){
    // db_type := ""
    // db_config := ""

    // var err error
    // DB_MySql, err = gorm.Open(db_type, db_config)
    // // defer DB.Close()
    // helper.CheckErr(err)

    // fmt.Println("----database--")

    // DB_MySql.DB().SetMaxIdleConns(10)   // 用于设置闲置的连接数
    // DB_MySql.DB().SetMaxOpenConns(100)  // 用于设置最大打开的连接数

    // DB_MySql.LogMode(true)  // 启用Logger，显示详细日志
}