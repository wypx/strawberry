package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func init() {
	db, _ := gorm.Open("sqlite3", "strawberry.db")
	defer db.Close()
}