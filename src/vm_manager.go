package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	HostAgent "vm_manager/host_agent"

	VmAdmin "vm_manager/vm_admin"
	VmAdmin2 "vm_manager/vm_admin2"

	VmAgent "vm_manager/vm_agent"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TelegramBot() {
	bot, err := tgbotapi.NewBotAPI("5042048090:AAH7QLPVYLhyp1ml6d4aoAQ7smlenx2qxU0")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			fmt.Println(runtime.GOROOT())      // GO 路径
			fmt.Println(runtime.Version())     //GO 版本信息 go1.9
			fmt.Println(os.Hostname())         //获得PC名
			fmt.Println(net.Interfaces())      //获得网卡信息
			fmt.Println(runtime.GOARCH)        //系统构架 386、amd64
			fmt.Println(runtime.GOOS)          //系统版本 windows
			fmt.Println(runtime.GOMAXPROCS(0)) //系统版本 windows

			bot.Send(msg)
		}
	}
}

func main() {
	// TelegramBot()
	log.Printf("vm manager start\n")
	VmAgent.Initialize()
	HostAgent.Initialize()
	// host_agent.Initialize()
	// vm_admin.Initialize()
	go func() {
		VmAdmin2.Initialize()
	}()

	VmAdmin.Initialize()
	log.Printf("vm manager exit\n")
}
