package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MobileStatus struct {
	RealMode    int
	SIMStatus   string
	SIMSignal   int
	DialStatus  string
	DialIPAddr  string
	DialIPMask  string
	DialGateway string
	DialDNSAddr string
}

const (
	NetTypeAuto = iota // 0
	NetType5G
	NetType4G
	NetType3G
	NetType2G
	numberOfNetType
)

const (
	DialTypeAuto = iota // 0
	DialTypeManual
	numberOfDialType
)

const (
	Sunday     = iota // 0
	Monday            // 1
	Tuesday           // 2
	Wedenesday        // 3
	Thursday          // 4
	Friday            // 5
	Saturday          // 6
	numberOfDays
)

/* Daily Dialing Plan */
type DailyPlan struct {
	StartHour   int
	StartMinute int
	StopHour    int
	StopMinute  int
}

type MobileConfig struct {
	/* Basic Config */
	NetType int
	Offline int
	User    string
	Pass    string
	APN     string
	MTU     int
	CHAP    int

	/* Dial Config */
	DialType int
	DialPlan map[int]interface{}

	/* SMS Config */
	EnableSMS   bool
	EnableAlarm bool
}

type Mobile struct {
	status MobileStatus
	config MobileConfig
}

// 用户登录页面
func (m *Mobile) Dial(c *gin.Context) {
	log.Printf("gin-login mobile")
	if c.Request.Method == "GET" {
		login := StawCtx.GetBool("gin-login")
		log.Printf("gin-login mobile: %t", login)
		if !login {
			RedirectLogin(c)
		} else {
			c.HTML(http.StatusOK, "strawberry_mobile", gin.H{
				"strawberry_title": "Mobile",
			})
		}
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")
		// http.StatusOK == 200
		c.JSON(http.StatusOK, gin.H{
			//"hello": session.Get("mysession"),
			"username": username,
			"password": password,
		})
	}
}
