package database

import (
	"vm_manager/vm_admin/helper"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func init() {
	db, err := gorm.Open("sqlite3", "/root/work/strawberry/bin/config/strawberry.db")
	helper.CheckErr(err)
	SetDB(db)
	// defer db.Close()
}
