package global

import (
	"github.com/jinzhu/gorm"
)

var (
	GlobalDB    *gorm.DB
	GlobalTimer Timer.Timer = Timer.NewTimerTask()
)
