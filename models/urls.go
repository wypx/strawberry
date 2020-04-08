package models

type Urls struct {
	ID   uint   `gorm:"primary_key"`
	Name string `gorm:"type:varchar(32);"`
	Url  string `gorm:"type:varchar(255);"`
}
