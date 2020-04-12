package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"tomato/helper"
)

func init() {
	db, err := gorm.Open("sqlite3", "strawberry.db")
	helper.CheckErr(err)
	SetDB(db)
	// defer db.Close()
}